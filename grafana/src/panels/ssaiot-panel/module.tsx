import { PanelPlugin } from '@grafana/data';
import { PanelOptions } from './types';
import { SSAIoTPanel } from './SSAIoTPanel';

export const plugin = new PanelPlugin<PanelOptions>(SSAIoTPanel).setPanelOptions((builder) => {
  return builder
    .addSelect({
      path: 'view',
      name: 'View',
      defaultValue: 'firmware-upload',
      settings: {
        options: [
          { value: 'firmware-upload', label: 'Firmware Uploader' },
          { value: 'firmware-browser', label: 'Firmware Browser' },
          { value: 'firmware-deploy', label: 'Firmware Deployer' },
        ],
      },
    })
    .addBooleanSwitch({
      path: 'defaultExpandAll',
      name: 'Expand all folders by default',
      defaultValue: true,
      showIf: (config) => config.view === 'firmware-browser',
    });
});
