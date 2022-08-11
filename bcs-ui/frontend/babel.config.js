

module.exports = function (api) {
    api.cache(true)

    const presets = [
        [
            '@babel/preset-env',
            {
                'modules': 'commonjs',
                'corejs': '3.1.4',
                'targets': {
                    'browsers': ['> 1%', 'last 2 versions', 'not ie <= 8']
                },
                'debug': false,
                'useBuiltIns': 'usage'
            }
        ],
        ['@vue/babel-preset-jsx']
    ]

    const plugins = [
        'lodash',
        '@babel/plugin-transform-runtime',
        '@babel/plugin-transform-async-to-generator',
        '@babel/plugin-transform-object-assign',
        '@babel/plugin-syntax-dynamic-import',
        'date-fns',
        '@vue/babel-plugin-transform-vue-jsx',
        '@babel/plugin-syntax-jsx',
        '@babel/plugin-proposal-optional-chaining',
        '@babel/plugin-proposal-class-properties'
    ]
    const comments = true

    return {
        compact: false,
        presets,
        plugins,
        comments,
        babelrcRoots: ['./src', './bk-bcs-saas/bcs-app/frontend/src/'],
        exclude: /node_modules/
    }
}
