/**
 * @param {object} options
 * @param {(params: { projectCode: string }) => Promise<Record<string, boolean>>} options.requestFlags
 * @param {(flags: Record<string, boolean>) => void} options.updateFeatureFlags
 * @param {Record<string, boolean>} options.defaultFlags
 */
function createFeatureFlagsLoader(options) {
  const { requestFlags, updateFeatureFlags, defaultFlags } = options;
  let latestRequestId = 0;

  /**
   * @param {{ projectCode: string }} params
   * @returns {Promise<Record<string, boolean>>}
   */
  return async function getFeatureFlags(params) {
    latestRequestId += 1;
    const requestId = latestRequestId;
    let data = {};
    try {
      data = await requestFlags(params) || {};
    } catch (error) {
      data = {};
    }
    if (requestId === latestRequestId) {
      updateFeatureFlags(Object.assign({}, defaultFlags, data));
    }
    return data;
  };
}

module.exports = {
  createFeatureFlagsLoader,
};
