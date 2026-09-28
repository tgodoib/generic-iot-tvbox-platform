import {
  DataSourceApi,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  TestDataSourceResponse,
} from '@grafana/data';

// Grafana's backendSrv throws a FetchError object (not a JS Error) on non-2xx.
function extractErrorMessage(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  if (err !== null && typeof err === 'object') {
    const e = err as Record<string, unknown>;
    const data = e['data'];
    if (typeof data === 'string') {
      return data;
    }
    if (data !== null && typeof data === 'object') {
      const d = data as Record<string, unknown>;
      return String(d['message'] ?? d['error'] ?? JSON.stringify(data));
    }
    return String(e['statusText'] ?? e['message'] ?? JSON.stringify(err));
  }
  return String(err);
}
import { getBackendSrv } from '@grafana/runtime';
import { lastValueFrom } from 'rxjs';
import { DataSourceOptions, DeviceClassItem, FirmwareListResponse, MyQuery } from './types';

export class DataSource extends DataSourceApi<MyQuery, DataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<DataSourceOptions>) {
    super(instanceSettings);
  }

  async query(_options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
    return { data: [] };
  }

  async testDatasource(): Promise<TestDataSourceResponse> {
    try {
      await lastValueFrom(
        getBackendSrv().fetch({
          url: `/api/datasources/uid/${this.uid}/resources/health`,
          method: 'GET',
          showErrorAlert: false,
        })
      );
      return { status: 'success', message: 'Connected successfully' };
    } catch (err: unknown) {
      return { status: 'error', message: `Connection failed: ${extractErrorMessage(err)}` };
    }
  }

  async listFirmwares(): Promise<DeviceClassItem[]> {
    const resp = await lastValueFrom(
      getBackendSrv().fetch<FirmwareListResponse>({
        url: `/api/datasources/uid/${this.uid}/resources/firmwares`,
        method: 'GET',
      })
    );
    return Object.entries(resp.data).map(([name, firmwares]) => ({
      name,
      firmwares: firmwares.map((f) => ({ version: f.version, lastModified: f.last_modified })),
    }));
  }

  async uploadFirmware(deviceClass: string, version: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('version', version);
    formData.append('device_class', deviceClass);
    await lastValueFrom(
      getBackendSrv().fetch({
        url: `/api/datasources/uid/${this.uid}/resources/firmwares`,
        method: 'POST',
        data: formData,
      })
    );
  }

  async sendUpdate(firmwareId: string, type: string, value: string): Promise<void> {
    await lastValueFrom(
      getBackendSrv().fetch({
        url: `/api/datasources/uid/${this.uid}/resources/update/${encodeURIComponent(firmwareId)}`,
        method: 'POST',
        data: { type, value },
      })
    );
  }

  async deleteFirmware(id: string): Promise<void> {
    await lastValueFrom(
      getBackendSrv().fetch({
        url: `/api/datasources/uid/${this.uid}/resources/firmware/${id}`,
        method: 'DELETE',
      })
    );
  }
}
