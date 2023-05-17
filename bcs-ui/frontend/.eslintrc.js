module.exports = {
    root: true,
    extends: ['@blueking/eslint-config-bk/tsvue'],
    rules: {
        "@typescript-eslint/no-misused-promises": [
            "error",
            {
                "checksVoidReturn": false
            }
        ]
    }
};