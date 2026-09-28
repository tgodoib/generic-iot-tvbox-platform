import React, { useState } from 'react';
import { PanelProps, GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { getBackendSrv, getAppEvents, RefreshEvent } from '@grafana/runtime';
import { useStyles2, Field, Input, FileUpload, Button, Alert } from '@grafana/ui';
import { lastValueFrom } from 'rxjs';
import { PanelOptions } from './types';

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    overflow-y: auto;
    padding: ${theme.spacing(2)};
    scrollbar-width: thin;
    scrollbar-color: ${theme.colors.border.medium} transparent;
    &::-webkit-scrollbar { width: 6px; }
    &::-webkit-scrollbar-track { background: transparent; }
    &::-webkit-scrollbar-thumb {
      background-color: ${theme.colors.border.medium};
      border-radius: 3px;
    }
  `,
  form: css`
    max-width: 480px;
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(2)};
  `,
  fileName: css`
    margin-top: ${theme.spacing(0.5)};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
  actions: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(2)};
    padding-top: ${theme.spacing(1)};
  `,
});

function extractError(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  if (err !== null && typeof err === 'object') {
    const e = err as Record<string, unknown>;
    const data = e['data'];
    let detail = '';
    if (typeof data === 'string' && data.length > 0) {
      detail = data;
    } else if (data !== null && typeof data === 'object') {
      const d = data as Record<string, unknown>;
      const msg = d['message'] ?? d['error'];
      if (msg !== undefined) {
        detail = String(msg);
      } else {
        const s = JSON.stringify(data);
        if (s !== '{}') { detail = s; }
      }
    }
    if (!detail) {
      detail = String(e['statusText'] ?? e['message'] ?? '');
    }
    const status = e['status'];
    if (status !== undefined) {
      return detail ? `[${status}] ${detail}` : `HTTP ${status}`;
    }
    return detail || JSON.stringify(err);
  }
  return String(err);
}

async function uploadFirmware(deviceClass: string, version: string, file: File): Promise<void> {
  const listResp = await lastValueFrom(
    getBackendSrv().fetch<Array<{ uid: string; type: string }>>({
      url: '/api/datasources',
      showErrorAlert: false,
    })
  );
  const ds = listResp.data.find((d) => d.type === 'tiagogb-ssaiot-datasource');
  if (!ds) {
    throw new Error('No SSAIoT datasource configured. Add one under Connections → Data sources.');
  }

  const formData = new FormData();
  formData.append('file', file);
  formData.append('version', version);
  formData.append('device_class', deviceClass);

  // Use native fetch — getBackendSrv().fetch() serialises `data` as JSON,
  // which drops the multipart boundary and sends Content-Type: application/json.
  // Same-origin requests include the Grafana session cookie automatically.
  const resp = await window.fetch(`/api/datasources/uid/${ds.uid}/resources/firmwares`, {
    method: 'POST',
    body: formData,
    credentials: 'same-origin',
  });

  if (!resp.ok) {
    let msg = `HTTP ${resp.status}`;
    try {
      const body = await resp.json() as Record<string, unknown>;
      const detail = body['message'] ?? body['error'];
      if (detail) { msg += `: ${detail}`; }
    } catch { /* non-JSON body, use status only */ }
    throw new Error(msg);
  }
}

export const FirmwareUpload: React.FC<PanelProps<PanelOptions>> = ({ width, height }) => {
  const styles = useStyles2(getStyles);

  const [deviceClass, setDeviceClass] = useState('');
  const [firmwareVersion, setFirmwareVersion] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [submitted, setSubmitted] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isValid = deviceClass.trim() !== '' && firmwareVersion.trim() !== '' && file !== null;

  const handleSubmit = async () => {
    if (!isValid || !file) {
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      await uploadFirmware(deviceClass.trim(), firmwareVersion.trim(), file);
      setSubmitted(true);
      getAppEvents().publish(new RefreshEvent());
    } catch (err: unknown) {
      setError(extractError(err));
    } finally {
      setIsLoading(false);
    }
  };

  const handleReset = () => {
    setDeviceClass('');
    setFirmwareVersion('');
    setFile(null);
    setSubmitted(false);
    setError(null);
  };

  return (
    <div style={{ width, height }} className={styles.container}>
      <div className={styles.form}>
        {submitted && (
          <Alert
            title="Firmware uploaded successfully"
            severity="success"
            onRemove={handleReset}
          >
            {file?.name} was uploaded to {deviceClass}/{firmwareVersion}.
          </Alert>
        )}

        {error && (
          <Alert title="Upload failed" severity="error" onRemove={() => setError(null)}>
            {error}
          </Alert>
        )}

        <Field label="Device Class" description="The device class this firmware belongs to">
          <Input
            id="device-class"
            placeholder="e.g. TemperatureSensor"
            value={deviceClass}
            onChange={(e) => setDeviceClass(e.currentTarget.value)}
            disabled={submitted}
          />
        </Field>

        <Field label="Firmware Version" description="Semantic version of the firmware">
          <Input
            id="firmware-version"
            placeholder="e.g. v2.1.0"
            value={firmwareVersion}
            onChange={(e) => setFirmwareVersion(e.currentTarget.value)}
            disabled={submitted}
          />
        </Field>

        <Field label="Firmware File" description="Binary firmware file to upload (.bin)">
          <div>
            {!submitted && (
              <FileUpload
                onFileUpload={(e) => {
                  const selected = e.currentTarget.files?.[0] ?? null;
                  setFile(selected);
                }}
              >
                {file ? 'Change file' : 'Select file'}
              </FileUpload>
            )}
            {file && (
              <div className={styles.fileName}>{file.name}</div>
            )}
          </div>
        </Field>

        <div className={styles.actions}>
          <Button
            variant="primary"
            icon={isLoading ? 'fa fa-spinner' : 'cloud-upload'}
            disabled={!isValid || submitted || isLoading}
            onClick={handleSubmit}
          >
            {isLoading ? 'Uploading…' : 'Upload Firmware'}
          </Button>
        </div>
      </div>
    </div>
  );
};
