module.exports = {
    root: true,
    extends: ['@blueking/eslint-config-bk/tsvue'],
    plugins: [
        'simple-import-sort',
    ],
    rules: {
        "@typescript-eslint/no-misused-promises": [
            "error",
            {
                "checksVoidReturn": false
            }
        ],
        'simple-import-sort/imports': ['error', {
            groups: [
              ['^[a-zA-Z]'],
              ['^@\\w'],
              ['^\\.\\.'],
              ['^\\.'],
            ],
        }],
        'vue/space-infix-ops': "off",
        "vue/multi-word-component-names": "off"
    }
};