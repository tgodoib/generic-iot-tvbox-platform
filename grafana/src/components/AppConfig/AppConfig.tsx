import React from 'react';
import { lastValueFrom } from 'rxjs';
import { AppPluginMeta, PluginConfigPageProps, PluginMeta } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';

export interface AppConfigProps extends PluginConfigPageProps<AppPluginMeta> {}

const AppConfig = ({ plugin }: AppConfigProps) => {
  const { enabled, pinned } = plugin.meta;

  // Persist enabled/pinned state on first load so Grafana shows the app as active.
  // All plugin configuration is done through the SSAIoT Datasource config editor.
  React.useEffect(() => {
    if (!enabled) {
      updatePlugin(plugin.meta.id, { enabled: true, pinned: pinned ?? true });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <p>
      This plugin has no app-level settings. Configure the server URL and mTLS certificates under{' '}
      <strong>Connections → Data sources → SSAIoT Datasource</strong>.
    </p>
  );
};

export default AppConfig;

const updatePlugin = async (pluginId: string, data: Partial<PluginMeta>) => {
  const response = getBackendSrv().fetch({
    url: `/api/plugins/${pluginId}/settings`,
    method: 'POST',
    data,
  });
  return lastValueFrom(response);
};
