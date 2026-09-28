import React from 'react';
import { DataSourcePluginOptionsEditorProps, GrafanaTheme2 } from '@grafana/data';
import { Button, InlineField, Input, TextArea, useStyles2 } from '@grafana/ui';
import { css } from '@emotion/css';
import { DataSourceOptions, SecureJsonData } from './types';

type Props = DataSourcePluginOptionsEditorProps<DataSourceOptions, SecureJsonData>;

const getStyles = (theme: GrafanaTheme2) => ({
  section: css`
    margin-bottom: ${theme.spacing(3)};
  `,
  heading: css`
    font-size: ${theme.typography.h5.fontSize};
    font-weight: ${theme.typography.fontWeightMedium};
    margin-bottom: ${theme.spacing(2)};
    color: ${theme.colors.text.primary};
  `,
  configuredField: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    height: 32px;
  `,
  configuredText: css`
    color: ${theme.colors.text.secondary};
    font-style: italic;
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
});

export function ConfigEditor({ options, onOptionsChange }: Props) {
  const styles = useStyles2(getStyles);

  const isClientCertConfigured = Boolean(options.secureJsonFields?.clientCert);
  const isClientKeyConfigured = Boolean(options.secureJsonFields?.clientKey);

  const onURLChange = (url: string) => {
    onOptionsChange({ ...options, url });
  };

  const onSecureChange = (field: keyof SecureJsonData, value: string) => {
    onOptionsChange({
      ...options,
      secureJsonData: { ...options.secureJsonData, [field]: value },
    });
  };

  const onSecureReset = (field: keyof SecureJsonData) => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...options.secureJsonFields, [field]: false },
      secureJsonData: { ...options.secureJsonData, [field]: '' },
    });
  };

  return (
    <div>
      <div className={styles.section}>
        <div className={styles.heading}>Connection</div>
        <InlineField label="Server URL" labelWidth={20} tooltip="Base URL of the SSAIoT firmware API">
          <Input
            value={options.url ?? ''}
            onChange={(e) => onURLChange(e.currentTarget.value)}
            placeholder="https://your-iot-server.example.com"
            width={50}
          />
        </InlineField>
      </div>

      <div className={styles.section}>
        <div className={styles.heading}>mTLS Client Authentication</div>

        <InlineField
          label="Client Certificate"
          labelWidth={20}
          tooltip="PEM-encoded client certificate for mTLS authentication"
        >
          {isClientCertConfigured ? (
            <div className={styles.configuredField}>
              <span className={styles.configuredText}>Certificate configured</span>
              <Button variant="secondary" size="sm" onClick={() => onSecureReset('clientCert')}>
                Reset
              </Button>
            </div>
          ) : (
            <TextArea
              value={options.secureJsonData?.clientCert ?? ''}
              onChange={(e) => onSecureChange('clientCert', e.currentTarget.value)}
              placeholder={'-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'}
              rows={6}
              cols={50}
            />
          )}
        </InlineField>

        <InlineField
          label="Client Key"
          labelWidth={20}
          tooltip="PEM-encoded private key for mTLS authentication"
        >
          {isClientKeyConfigured ? (
            <div className={styles.configuredField}>
              <span className={styles.configuredText}>Key configured</span>
              <Button variant="secondary" size="sm" onClick={() => onSecureReset('clientKey')}>
                Reset
              </Button>
            </div>
          ) : (
            <TextArea
              value={options.secureJsonData?.clientKey ?? ''}
              onChange={(e) => onSecureChange('clientKey', e.currentTarget.value)}
              placeholder={'-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'}
              rows={6}
              cols={50}
            />
          )}
        </InlineField>
      </div>
    </div>
  );
}
