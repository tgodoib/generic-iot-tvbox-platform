import React from 'react';
import { render, screen } from '@testing-library/react';
import { PluginType } from '@grafana/data';
import AppConfig, { AppConfigProps } from './AppConfig';

describe('Components/AppConfig', () => {
  let props: AppConfigProps;

  beforeEach(() => {
    jest.resetAllMocks();

    props = {
      plugin: {
        meta: {
          id: 'sample-app',
          name: 'Sample App',
          type: PluginType.app,
          enabled: true,
          jsonData: {},
        },
      },
      query: {},
    } as unknown as AppConfigProps;
  });

  test('renders the datasource configuration note', () => {
    // @ts-ignore - We don't need to provide `addConfigPage()` and `setChannelSupport()` for these tests
    render(<AppConfig plugin={props.plugin} query={props.query} />);

    expect(screen.getByText(/configure the server url and mtls certificates/i)).toBeInTheDocument();
  });
});
