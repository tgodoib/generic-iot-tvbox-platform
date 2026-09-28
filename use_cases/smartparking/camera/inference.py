import ast
import csv
import io
import logging
import os
import random
import subprocess
from dataclasses import dataclass
from typing import Optional

import numpy as np
# import picamera
from numpy import sqrt
from PIL import Image, ImageDraw
from ai_edge_litert.interpreter import Interpreter

logger = logging.getLogger(__name__)


@dataclass
class InferenceResult:
    parking_status: str
    occupied_count: int
    total_spots: int
    bitmask: int
    spots: list
    annotated_image: Optional[Image.Image] = None
    raw_image: Optional[Image.Image] = None


class SmartParkingCamera:
    def __init__(
            self,
            model_path: str,
            coordinates_csv_path: str,
            mask_path: str,
            confidence_threshold: float = 0.25,
            max_assignment_distance: float = 0.1,
            large_detection_spot_ids: Optional[list] = None,
            large_detection_area_threshold: float = 4000.0,
    ):
        self.confidence_threshold = confidence_threshold
        self.max_assignment_distance = max_assignment_distance
        self.large_detection_spot_ids = large_detection_spot_ids if large_detection_spot_ids is not None else [5, 6, 7]
        self.large_detection_area_threshold = large_detection_area_threshold

        self.spots = self._load_spot_coordinates(coordinates_csv_path)
        self.mask = self._load_mask(mask_path)
        (
            self._interpreter,
            self._input_details,
            self._output_details,
            self._model_width,
            self._model_height,
            self._model_dtype,
        ) = self._load_interpreter(model_path)
        logger.info(f"Loaded {len(self.spots)} spots. Model input: {self._model_width}x{self._model_height}, dtype {self._model_dtype.__name__}.")

    def _load_spot_coordinates(self, csv_path: str) -> list:
        try:
            with open(csv_path, newline="", encoding="utf-8") as f:
                reader = csv.DictReader(f)
                rows = list(reader)
                if not rows:
                    return []
                spot_boxes = ast.literal_eval(rows[0]["yolo_boxes"])
                spots = []
                for spot_id, box in enumerate(spot_boxes):
                    if len(box) >= 3:
                        spots.append({
                            "id": spot_id,
                            "x_center": box[1],
                            "y_center": box[2],
                            "width": box[3] if len(box) > 3 else 0.05,
                            "height": box[4] if len(box) > 4 else 0.05,
                            "occupied": False,
                        })
                return spots
        except Exception as e:
            logger.error(f"Error loading spot coordinates: {e}")
            return []

    def _load_mask(self, mask_path: str) -> np.ndarray:
        return np.array(Image.open(mask_path))

    def _load_interpreter(self, model_path: str):
        interpreter = Interpreter(model_path, num_threads=4)
        interpreter.allocate_tensors()
        input_details = interpreter.get_input_details()
        output_details = interpreter.get_output_details()
        _, width, height, _ = input_details[0]["shape"]
        dtype = input_details[0]["dtype"]
        return interpreter, input_details, output_details, width, height, dtype

    def _capture_image(self) -> Optional[Image.Image]:
        # ===== Cycles through demo images =====
        demo_dir = "./assets/demo_images"
        images = os.listdir(demo_dir)
        if not images:
            logger.error("No demo images found in ./assets/demo_images.")
            return None
        return Image.open(os.path.join(demo_dir, random.choice(images))).copy()

        # ===== Gets the most recent image from TotemServer =====
        # local_path = "./assets/last_pic.jpg"
        # cmd = (
        #         "rsync -avz "
        #         "\"root@100.121.208.3:"
        #         "$(ssh root@100.121.208.3 "
        #         "'ls /home/unicamp/photos_pi/ | grep -E \"^[0-9]\" | tail -1 | xargs -I{} echo /home/unicamp/photos_pi/{}')\" "
        #         + local_path
        # )
        #
        # try:
        #     subprocess.run(cmd, shell=True, check=True)
        # except subprocess.CalledProcessError as e:
        #     logger.error(f"Failed to fetch image from remote: {e}")
        #     return None
        # return Image.open(local_path).copy()

        # ===== Takes a picture =====
        # stream = io.BytesIO()
        # with picamera.PiCamera(resolution="1024x768") as camera:
        #     camera.rotation = 180
        #     camera.start_preview()
        #     time.sleep(2)
        #     camera.capture(stream, format="jpeg")
        # stream.seek(0)
        # return Image.open(stream).copy()

    def _process_image(self, img: Image.Image) -> np.ndarray:
        img = img.resize((self._model_width, self._model_height))
        arr = np.expand_dims(np.array(img), axis=0)
        if "float" in str(self._model_dtype):
            arr = (arr / 256.0).astype(self._model_dtype)
        return arr

    def _detect(self, input_data: np.ndarray) -> np.ndarray:
        self._interpreter.set_tensor(self._input_details[0]["index"], input_data)
        self._interpreter.invoke()
        output = self._interpreter.get_tensor(self._output_details[0]["index"])

        detected_boxes = []
        detected_scores = []
        if self._output_details[0]["name"] == "Identity":
            output = output[0].T
            scores = np.max(output[..., 4:], axis=1)
            classes = np.argmax(output[..., 4:], axis=1)
            boxes_xywh = output[..., :4]
            for i in range(len(scores)):
                if scores[i] >= self.confidence_threshold and classes[i] in (2, 7):
                    detected_boxes.append(boxes_xywh[i])
                    detected_scores.append(scores[i])
        else:
            for det in output[0]:
                if det[-1] == 1 and det[-2] >= self.confidence_threshold:
                    detected_boxes.append(det[:4])
                    detected_scores.append(det[-2])

        if not detected_boxes:
            return np.empty((0, 4))

        detected_boxes = np.array(detected_boxes)
        detected_scores = np.array(detected_scores)
        keep = self._nms(detected_boxes, scores=detected_scores)
        return detected_boxes[keep]

    def _nms(self, boxes: np.ndarray, scores: np.ndarray, iou_threshold: float = 0.5) -> list:
        x, y, w, h = boxes[:, 0], boxes[:, 1], boxes[:, 2], boxes[:, 3]
        areas = w * h
        order = scores.argsort()[::-1]
        keep = []
        while order.size > 0:
            i = order[0]
            keep.append(i)
            xx1 = np.maximum(x[i], x[order[1:]])
            yy1 = np.maximum(y[i], y[order[1:]])
            xx2 = np.minimum(x[i] + w[i], x[order[1:]] + w[order[1:]])
            yy2 = np.minimum(y[i] + h[i], y[order[1:]] + h[order[1:]])
            inter = np.maximum(0, xx2 - xx1) * np.maximum(0, yy2 - yy1)
            iou = inter / (areas[i] + areas[order[1:]] - inter)
            order = order[np.where(iou <= iou_threshold)[0] + 1]
        return keep

    def _is_inside_mask(self, x: float, y: float) -> bool:
        channel = self.mask[:, :, 0]
        mask_h, mask_w = channel.shape
        return channel[int(y * mask_h), int(x * mask_w)] != 255

    def _split_large_detection(self, det, img_width: int, img_height: int) -> list:
        x_center, y_center, w, h = det
        if (w * img_width) * (h * img_height) <= self.large_detection_area_threshold:
            return [det]
        logger.debug(f"Splitting large detection at ({x_center:.3f}, {y_center:.3f}), area={(w * img_width * h * img_height):.0f}px².")
        half_w = w / 2
        return [
            [x_center - half_w / 2, y_center, half_w, h],
            [x_center + half_w / 2, y_center, half_w, h],
        ]

    def _split_boxes_near_spots(self, boxes: np.ndarray, img_width: int, img_height: int) -> np.ndarray:
        if len(boxes) == 0:
            return boxes
        new_boxes = []
        for det in boxes:
            for s in self.spots:
                if s["id"] in self.large_detection_spot_ids:
                    dist = sqrt((det[0] - s["x_center"]) ** 2 + (det[1] - s["y_center"]) ** 2)
                    if dist < self.max_assignment_distance:
                        new_boxes.extend(self._split_large_detection(det, img_width, img_height))
                        break
            else:
                new_boxes.append(det)
        return np.array(new_boxes)

    def _assign_detections_to_spots(self, boxes: np.ndarray, img_width: int, img_height: int):
        for spot in self.spots:
            spot["occupied"] = False

        assignments = []
        for det_idx, det in enumerate(boxes):
            if len(det) < 2:
                continue
            x, y = det[0], det[1]
            nearest_idx, min_dist = None, float("inf")
            for spot_idx, spot in enumerate(self.spots):
                dist = sqrt((x - spot["x_center"]) ** 2 + (y - spot["y_center"]) ** 2)
                if dist < min_dist and dist < self.max_assignment_distance:
                    min_dist, nearest_idx = dist, spot_idx
            if nearest_idx is not None and self._is_inside_mask(x, y):
                self.spots[nearest_idx]["occupied"] = True
                assignments.append((det_idx, nearest_idx))

        return self.spots, assignments

    def _draw_parking_status(
            self, image: Image.Image, boxes: np.ndarray, assignments: list, img_width: int, img_height: int
    ) -> Image.Image:
        draw = ImageDraw.Draw(image)
        for spot in self.spots:
            x = spot["x_center"] * img_width
            y = spot["y_center"] * img_height
            w = spot["width"] * img_width / 2
            h = spot["height"] * img_height / 2
            color = (255, 0, 0, 150) if spot["occupied"] else (0, 255, 0, 150)
            draw.rectangle([x - w, y - h, x + w, y + h], outline=color, width=2)
            draw.text((x - w + 5, y - h + 5), f"Spot {spot['id']}", fill="white")

        for det_idx, spot_idx in assignments:
            det = boxes[det_idx]
            spot = self.spots[spot_idx]
            x1 = (det[0] - det[2] / 2) * img_width
            y1 = (det[1] - det[3] / 2) * img_height
            x2 = (det[0] + det[2] / 2) * img_width
            y2 = (det[1] + det[3] / 2) * img_height
            draw.rectangle([x1, y1, x2, y2], outline="yellow", width=2)
            draw.line(
                [(det[0] * img_width, det[1] * img_height),
                 (spot["x_center"] * img_width, spot["y_center"] * img_height)],
                fill="cyan",
                width=2,
            )
        return image

    def run_inference(self) -> Optional[InferenceResult]:
        img = self._capture_image()
        if img is None:
            return None
        img_width, img_height = img.size
        # logger.info(f"Image captured: {img_width}x{img_height}")

        boxes = self._detect(self._process_image(img))
        boxes = self._split_boxes_near_spots(boxes, img_width, img_height)
        spots, assignments = self._assign_detections_to_spots(boxes, img_width, img_height)

        parking_status = "".join(
            "1" if s["occupied"] else "0"
            for s in sorted(spots, key=lambda s: s["id"])
        )
        occupied_count = parking_status.count("1")
        bitmask = int(parking_status, 2) if parking_status else 0

        logger.debug(f"Parking status: {parking_status} ({occupied_count}/{len(spots)} occupied)")

        raw = img.copy()
        annotated = self._draw_parking_status(img.copy(), boxes, assignments, img_width, img_height)

        return InferenceResult(
            parking_status=parking_status,
            occupied_count=occupied_count,
            total_spots=len(spots),
            bitmask=bitmask,
            spots=spots,
            annotated_image=annotated,
            raw_image=raw,
        )
