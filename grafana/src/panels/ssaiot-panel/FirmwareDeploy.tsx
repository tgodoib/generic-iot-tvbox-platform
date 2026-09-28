import React, { useState } from 'react';
import { PanelProps, GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { getBackendSrv } from '@grafana/runtime';
import { useStyles2, Field, Input, Button, Alert, RadioButtonGroup } from '@grafana/ui';
import { lastValueFrom } from 'rxjs';
import { PanelOptions } from './types';

type TargetMode = 'class' | 'device';

const TARGET_OPTIONS = [
  { label: 'Device Class', value: 'class' as TargetMode },
  { label: 'Device ID', value: 'device' as TargetMode },
];

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

async function sendUpdate(firmwareId: string, type: string, value: string): Promise<void> {
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

  await lastValueFrom(
    getBackendSrv().fetch({
      // firmwareId may contain slashes (e.g. "esp32/1.0.0") — pass as path segments, not encoded.
      url: `/api/datasources/uid/${ds.uid}/resources/update/${firmwareId}`,
      method: 'POST',
      data: { type, value },
      showErrorAlert: false,
    })
  );
}

export const FirmwareDeploy: React.FC<PanelProps<PanelOptions>> = ({ width, height }) => {
  const styles = useStyles2(getStyles);

  const [firmwareId, setFirmwareId] = useState('');
  const [targetMode, setTargetMode] = useState<TargetMode>('class');
  const [targetValue, setTargetValue] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isValid = firmwareId.trim() !== '' && targetValue.trim() !== '';

  const handleSubmit = async () => {
    if (!isValid) {
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      // API expects type "class" or "id", not the internal "device" label.
      const apiType = targetMode === 'class' ? 'class' : 'id';
      await sendUpdate(firmwareId.trim(), apiType, targetValue.trim());
      setSubmitted(true);
    } catch (err: unknown) {
      setError(extractError(err));
    } finally {
      setIsLoading(false);
    }
  };

  const handleReset = () => {
    setFirmwareId('');
    setTargetValue('');
    setSubmitted(false);
    setError(null);
  };

  const targetLabel = targetMode === 'class' ? 'Device Class' : 'Device ID';
  const targetPlaceholder = targetMode === 'class' ? 'e.g. TemperatureSensor' : 'e.g. device-00a1b2c3';
  const targetDescription =
    targetMode === 'class'
      ? 'Deploy to all devices belonging to this class'
      : 'Deploy to a single device by its ID';

  return (
    <div style={{ width, height }} className={styles.container}>
      <div className={styles.form}>
        {submitted && (
          <Alert
            title="Update dispatched successfully"
            severity="success"
            onRemove={handleReset}
          >
            Firmware <strong>{firmwareId}</strong> was sent to{' '}
            {targetMode === 'class' ? 'class' : 'device'} <strong>{targetValue}</strong>.
          </Alert>
        )}

        {error && (
          <Alert title="Deploy failed" severity="error" onRemove={() => setError(null)}>
            {error}
          </Alert>
        )}

        <Field
          label="Firmware ID"
          description="Firmware to deploy, in the format DeviceClass/version"
        >
          <Input
            id="firmware-id"
            placeholder="e.g. esp32/1.0.0"
            value={firmwareId}
            onChange={(e) => setFirmwareId(e.currentTarget.value)}
            disabled={submitted}
          />
        </Field>

        <Field label="Target Type" description="Deploy to an entire device class or a single device">
          <RadioButtonGroup
            options={TARGET_OPTIONS}
            value={targetMode}
            onChange={(v) => {
              setTargetMode(v);
              setTargetValue('');
            }}
            disabled={submitted}
          />
        </Field>

        <Field label={targetLabel} description={targetDescription}>
          <Input
            id="target-value"
            placeholder={targetPlaceholder}
            value={targetValue}
            onChange={(e) => setTargetValue(e.currentTarget.value)}
            disabled={submitted}
          />
        </Field>

        <div className={styles.actions}>
          <Button
            variant="primary"
            icon={isLoading ? 'fa fa-spinner' : 'sync'}
            disabled={!isValid || submitted || isLoading}
            onClick={handleSubmit}
          >
            {isLoading ? 'Dispatching…' : 'Deploy Firmware'}
          </Button>
        </div>
      </div>
    </div>
  );
};
