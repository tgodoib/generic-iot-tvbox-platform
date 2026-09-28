import pluginJson from './plugin.json';

export const PLUGIN_BASE_URL = `/a/${pluginJson.id}`;

export enum ROUTES {
  FirmwareBrowser = 'firmware-browser',
  FirmwareUpload = 'firmware-upload',
}
