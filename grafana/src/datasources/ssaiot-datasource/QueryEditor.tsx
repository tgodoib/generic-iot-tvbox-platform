import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './DataSource';
import { DataSourceOptions, MyQuery } from './types';

type Props = QueryEditorProps<DataSource, MyQuery, DataSourceOptions>;

export function QueryEditor(_props: Props) {
  return null;
}
