module.exports = {
    root: true,
    extends: ['@blueking/eslint-config-bk/tsvue'],
    plugins: [
        'header',
    ],
    overrides: [
    {
        files: ['*.js'],
        rules: {
            'header/header': [2, './config/header.js'],
        },
    },
    ],
};