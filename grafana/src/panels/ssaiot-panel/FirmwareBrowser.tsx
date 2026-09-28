import React, { useState, useMemo, useEffect } from 'react';
import { PanelProps, GrafanaTheme2 } from '@grafana/data';
import { css, cx } from '@emotion/css';
import { getBackendSrv } from '@grafana/runtime';
import { useStyles2, Icon, Input, IconButton, Spinner, Alert, ConfirmModal } from '@grafana/ui';
import { lastValueFrom } from 'rxjs';
import { PanelOptions } from './types';

interface FirmwareItem {
  version: string;
  lastModified: string;
}

interface DeviceClassItem {
  name: string;
  firmwares: FirmwareItem[];
}

type FirmwareListResponse = Record<string, Array<{ version: string; last_modified: string }>>;

const fmtDate = (iso: string) => iso.replace('T', ' ').slice(0, 19);

function extractError(err: unknown): string {
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

async function findDatasourceUID(): Promise<string> {
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
  return ds.uid;
}

async function fetchFirmwares(): Promise<DeviceClassItem[]> {
  const uid = await findDatasourceUID();
  const resp = await lastValueFrom(
    getBackendSrv().fetch<FirmwareListResponse>({
      url: `/api/datasources/uid/${uid}/resources/firmwares`,
      method: 'GET',
      showErrorAlert: false,
    })
  );
  return Object.entries(resp.data).map(([name, firmwares]) => ({
    name,
    firmwares: firmwares.map((f) => ({ version: f.version, lastModified: f.last_modified })),
  }));
}

async function deleteFirmware(id: string): Promise<void> {
  const uid = await findDatasourceUID();
  await lastValueFrom(
    getBackendSrv().fetch({
      url: `/api/datasources/uid/${uid}/resources/firmware/${id}`,
      method: 'DELETE',
      showErrorAlert: false,
    })
  );
}

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    overflow-y: auto;
    padding: ${theme.spacing(1)};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    scrollbar-width: thin;
    scrollbar-color: ${theme.colors.border.medium} transparent;
    &::-webkit-scrollbar { width: 6px; }
    &::-webkit-scrollbar-track { background: transparent; }
    &::-webkit-scrollbar-thumb {
      background-color: ${theme.colors.border.medium};
      border-radius: 3px;
    }
    &::-webkit-scrollbar-thumb:hover {
      background-color: ${theme.colors.border.strong};
    }
  `,
  searchWrapper: css`
    margin-bottom: ${theme.spacing(1)};
  `,
  folder: css`
    margin-bottom: ${theme.spacing(0.5)};
    border-radius: ${theme.shape.radius.default};
    overflow: hidden;
    border: 1px solid ${theme.colors.border.weak};
  `,
  folderHeader: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    padding: ${theme.spacing(0.75, 1)};
    background: ${theme.colors.background.secondary};
    cursor: pointer;
    user-select: none;
    color: ${theme.colors.text.primary};
    &:hover { background: ${theme.colors.action.hover}; }
  `,
  folderName: css`
    flex: 1;
    font-weight: ${theme.typography.fontWeightMedium};
    color: ${theme.colors.text.primary};
  `,
  fileCount: css`
    color: ${theme.colors.text.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  fileList: css`
    background: ${theme.colors.background.primary};
  `,
  fileRow: css`
    display: grid;
    grid-template-columns: auto 1fr auto auto auto;
    align-items: center;
    gap: ${theme.spacing(1)};
    padding: ${theme.spacing(0.5, 1, 0.5, 3)};
    color: ${theme.colors.text.secondary};
    border-top: 1px solid ${theme.colors.border.weak};
    &:hover {
      background: ${theme.colors.action.hover};
      color: ${theme.colors.text.primary};
    }
  `,
  fileName: css`
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  `,
  fileMeta: css`
    color: ${theme.colors.text.disabled};
    font-size: ${theme.typography.bodySmall.fontSize};
    white-space: nowrap;
  `,
  chevron: css`
    transition: transform 0.15s ease;
    color: ${theme.colors.text.secondary};
  `,
  chevronOpen: css`
    transform: rotate(90deg);
  `,
  folderIcon: css`
    color: ${theme.colors.warning.text};
  `,
  fileIcon: css`
    color: ${theme.colors.text.disabled};
  `,
  highlight: css`
    background-color: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.contrastText};
    border-radius: 2px;
    padding: 0 2px;
  `,
});

interface PendingDelete {
  deviceClass: string;
  version: string;
}

export const FirmwareBrowser: React.FC<PanelProps<PanelOptions>> = ({ options, width, height, renderCounter }) => {
  const styles = useStyles2(getStyles);

  const [data, setData] = useState<DeviceClassItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [openFolders, setOpenFolders] = useState<Record<string, boolean>>({});
  const [search, setSearch] = useState('');
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [pendingDelete, setPendingDelete] = useState<PendingDelete | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const loadFirmwares = (cancelled: { value: boolean }) => {
    setLoading(true);
    setError(null);
    fetchFirmwares()
      .then((result) => {
        if (cancelled.value) { return; }
        setData(result);
        setOpenFolders((prev) => {
          const next: Record<string, boolean> = {};
          for (const d of result) {
            next[d.name] = d.name in prev ? prev[d.name] : options.defaultExpandAll;
          }
          return next;
        });
      })
      .catch((err: unknown) => {
        if (cancelled.value) { return; }
        setError(extractError(err));
      })
      .finally(() => {
        if (!cancelled.value) { setLoading(false); }
      });
  };

  useEffect(() => {
    const cancelled = { value: false };
    loadFirmwares(cancelled);
    return () => { cancelled.value = true; };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [renderCounter]);

  const handleDeleteConfirm = async () => {
    if (!pendingDelete) { return; }
    const { deviceClass, version } = pendingDelete;
    setPendingDelete(null);
    setIsDeleting(true);
    try {
      await deleteFirmware(`${deviceClass}/${version}`);
      const cancelled = { value: false };
      loadFirmwares(cancelled);
    } catch (err: unknown) {
      setError(extractError(err));
      setLoading(false);
    } finally {
      setIsDeleting(false);
    }
  };

  const toggle = (name: string) => {
    setOpenFolders((prev) => ({ ...prev, [name]: !prev[name] }));
  };

  const copyToClipboard = (deviceClass: string, firmware: string) => {
    const text = `${deviceClass}/${firmware}`;
    navigator.clipboard.writeText(text).then(() => {
      setCopiedKey(text);
      setTimeout(() => setCopiedKey(null), 2000);
    });
  };

  const query = search.trim().toLowerCase();

  const filteredData = useMemo(() => {
    if (!query) {
      return data;
    }
    return data.flatMap((deviceClass) => {
      const classMatches = deviceClass.name.toLowerCase().includes(query);
      const matchingFirmwares = deviceClass.firmwares.filter((fw) =>
        fw.version.toLowerCase().includes(query)
      );
      if (classMatches) {
        return [deviceClass];
      }
      if (matchingFirmwares.length > 0) {
        return [{ ...deviceClass, firmwares: matchingFirmwares }];
      }
      return [];
    });
  }, [query, data]);

  const highlight = (text: string) => {
    if (!query) {
      return <>{text}</>;
    }
    const idx = text.toLowerCase().indexOf(query);
    if (idx === -1) {
      return <>{text}</>;
    }
    return (
      <>
        {text.slice(0, idx)}
        <mark className={styles.highlight}>{text.slice(idx, idx + query.length)}</mark>
        {text.slice(idx + query.length)}
      </>
    );
  };

  if (loading) {
    return (
      <div style={{ width, height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spinner />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ width, height, padding: 16 }}>
        <Alert title="Failed to load firmwares" severity="error">{error}</Alert>
      </div>
    );
  }

  return (
    <div className={cx(styles.container, css`width: ${width}px; height: ${height}px;`)}>
      <div className={styles.searchWrapper}>
        <Input
          placeholder="Search device classes or firmwares…"
          prefix={<Icon name="search" />}
          value={search}
          onChange={(e) => setSearch(e.currentTarget.value)}
        />
      </div>
      {filteredData.map((deviceClass) => {
        const isOpen = query ? true : openFolders[deviceClass.name];
        return (
          <div key={deviceClass.name} className={styles.folder}>
            <div className={styles.folderHeader} onClick={() => !query && toggle(deviceClass.name)}>
              <Icon name="angle-right" className={cx(styles.chevron, isOpen && styles.chevronOpen)} size="sm" />
              <Icon name={isOpen ? 'folder-open' : 'folder'} className={styles.folderIcon} size="sm" />
              <span className={styles.folderName}>{highlight(deviceClass.name)}</span>
              <span className={styles.fileCount}>{deviceClass.firmwares.length} firmwares</span>
            </div>
            {isOpen && (
              <div className={styles.fileList}>
                {deviceClass.firmwares.map((fw) => {
                  const copyKey = `${deviceClass.name}/${fw.version}`;
                  const isCopied = copiedKey === copyKey;
                  return (
                    <div key={fw.version} className={styles.fileRow}>
                      <Icon name="file-alt" className={styles.fileIcon} size="sm" />
                      <span className={styles.fileName}>{highlight(fw.version)}</span>
                      <span className={styles.fileMeta}>{fmtDate(fw.lastModified)}</span>
                      <IconButton
                        name={isCopied ? 'check' : 'copy'}
                        tooltip={isCopied ? 'Copied!' : `Copy ${copyKey}`}
                        size="sm"
                        onClick={() => copyToClipboard(deviceClass.name, fw.version)}
                        aria-label="Copy firmware path"
                      />
                      <IconButton
                        name="trash-alt"
                        tooltip={`Delete ${copyKey}`}
                        size="sm"
                        variant="destructive"
                        disabled={isDeleting}
                        onClick={() => setPendingDelete({ deviceClass: deviceClass.name, version: fw.version })}
                        aria-label="Delete firmware"
                      />
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        );
      })}

      <ConfirmModal
        isOpen={pendingDelete !== null}
        title="Delete firmware"
        body={
          pendingDelete
            ? `Are you sure you want to delete ${pendingDelete.deviceClass}/${pendingDelete.version}? This cannot be undone.`
            : ''
        }
        confirmText="Delete"
        confirmButtonVariant="destructive"
        onConfirm={handleDeleteConfirm}
        onDismiss={() => setPendingDelete(null)}
      />
    </div>
  );
};
