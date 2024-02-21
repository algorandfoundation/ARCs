import terser from '@rollup/plugin-terser'
import typescript from '@rollup/plugin-typescript'
import dts from 'rollup-plugin-dts'
import esbuild from 'rollup-plugin-esbuild'

export default [
  {
    input: 'src/index.ts',
    output: [
      {
        file: 'dist/index.cjs.js',
        format: 'cjs',
        sourcemap: true
      },
      {
        file: 'dist/index.esm.js',
        format: 'esm',
        sourcemap: true
      }
    ],
    plugins: [
      typescript({ tsconfig: './tsconfig.build.json' }),
      process.env.NODE_ENV === 'production' ? terser() : esbuild()
    ]
  },
  {
    input: 'src/index.ts',
    output: [{
      file: 'dist/index.d.ts',
      format: 'esm'
    }],
    plugins: [dts()]
  }
]
