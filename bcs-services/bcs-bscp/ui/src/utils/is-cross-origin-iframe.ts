export default function isCrossOriginIFrame () {
  try {
    // @ts-ignore
    return !window.top.location.hostname
  } catch (e) {
    return true
  }
}