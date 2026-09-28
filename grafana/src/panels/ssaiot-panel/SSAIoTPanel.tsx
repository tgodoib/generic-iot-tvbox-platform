import React from 'react';
import { PanelProps } from '@grafana/data';
import { PanelOptions } from './types';
import { FirmwareBrowser } from './FirmwareBrowser';
import { FirmwareUpload } from './FirmwareUpload';
import { FirmwareDeploy } from './FirmwareDeploy';

export const SSAIoTPanel: React.FC<PanelProps<PanelOptions>> = (props) => {
  switch (props.options.view) {
    case 'firmware-upload':
      return <FirmwareUpload {...props} />;
    case 'firmware-deploy':
      return <FirmwareDeploy {...props} />;
    case 'firmware-browser':
    default:
      return <FirmwareBrowser {...props} />;
  }
};
