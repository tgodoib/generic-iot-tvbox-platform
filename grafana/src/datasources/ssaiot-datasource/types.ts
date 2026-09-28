import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {}

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface DataSourceOptions extends DataSourceJsonData {}

export interface SecureJsonData {
  clientCert?: string;
  clientKey?: string;
}

// Shape returned by GET /firmware on the server
export interface FirmwareEntry {
  version: string;
  last_modified: string;
}
export type FirmwareListResponse = Record<string, FirmwareEntry[]>;

// Normalized internal representation used by the panel
export interface FirmwareItem {
  version: string;
  lastModified: string;
}

export interface DeviceClassItem {
  name: string;
  firmwares: FirmwareItem[];
}
