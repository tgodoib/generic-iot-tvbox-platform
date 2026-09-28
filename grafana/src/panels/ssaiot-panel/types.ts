export type PanelView = 'firmware-browser' | 'firmware-upload' | 'firmware-deploy';

export interface PanelOptions {
  view: PanelView;
  defaultExpandAll: boolean;
}
