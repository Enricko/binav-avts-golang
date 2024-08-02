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

function formatDateTime(input) {
  const date = new Date(input);

  const options = {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour12: false,
    timeZone: "Asia/Jakarta",
  };

  const formattedDate = date.toLocaleString("en-GB", options).replace(",", "");
  return formattedDate.replace(/\//g, "-");
}
