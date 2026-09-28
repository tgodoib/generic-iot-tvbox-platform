import type { Configuration } from 'webpack';
import { merge } from 'webpack-merge';
import grafanaConfig, { Env } from './.config/webpack/webpack.config';
import CopyWebpackPlugin from 'copy-webpack-plugin';

const config = async (env: Env): Promise<Configuration> => {
  const baseConfig = await grafanaConfig(env);

  return merge(baseConfig, {
    plugins: [
      new CopyWebpackPlugin({
        patterns: [
          {
            from: 'panels/**/img/**',
            to: '.',
            noErrorOnMissing: true,
          },
          {
            from: 'datasources/**/img/**',
            to: '.',
            noErrorOnMissing: true,
          },
        ],
      }),
    ],
  });
};

export default config;
