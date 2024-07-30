function getBaseURLHost() {
    // Get the full URL
    const fullURL = window.location.href;
    // Get the protocol and hostname
    const host = window.location.host;
    // Combine them to get the base URL
    const baseURL = `${host}`;
    return baseURL;
}
const websocketUrl = `ws://${getBaseURLHost()}/ws/kapal`;

function getBaseURL() {
    // Get the full URL
    const fullURL = window.location.href;
    // Get the protocol and hostname
    const protocol = window.location.protocol;
    const host = window.location.host;
    // Combine them to get the base URL
    const baseURL = `${protocol}//${host}/`;
    return baseURL;
  }