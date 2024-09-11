function getBaseURLHost() {
  // Get the full URL
  const fullURL = window.location.href;
  // Get the protocol and hostname
  const host = window.location.host;
  // Combine them to get the base URL
  const baseURL = `${host}`;
  return baseURL;
}
const websocketUrl = `ws://${getBaseURLHost()}/ws`;

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
    timeZone: "UTC",
  };

  const formattedDate = date.toLocaleString("en-GB", options).replace(",", "");
  return formattedDate.replace(/\//g, "-") + " UTC";
}

function formatDateTimeDisplay(input) {
  const date = new Date(input);

  const options = {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour12: false,
    timeZone: "UTC",
  };

  const formattedDate = date.toLocaleString("en-GB", options).replace(",", "");
  return formattedDate + " UTC";
}

const formatDate = (date) => {
  const pad = (num) => String(num).padStart(2, "0");
  const year = date.getUTCFullYear();
  const month = pad(date.getUTCMonth() + 1);
  const day = pad(date.getUTCDate());
  const hours = pad(date.getUTCHours());
  const minutes = pad(date.getUTCMinutes());
  return `${year}-${month}-${day}T${hours}:${minutes}`;
};

const formatDateWithMidnight = (date) => {
  const pad = (num) => String(num).padStart(2, "0");
  const year = date.getUTCFullYear();
  const month = pad(date.getUTCMonth() + 1);
  const day = pad(date.getUTCDate());
  return `${year}-${month}-${day}T00:00`;
};

function startToEndDatetimeFilterForm() {
  // Get current date and time in UTC
  const now = new Date();

  // Set default end date to now in UTC
  const endDateTime = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate(), now.getUTCHours(), now.getUTCMinutes()));

  // Set default start date to 3 days ago with time 00:00:00 UTC
  const startDateTime = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - 3, 0, 0, 0));

  // Try to set input values if elements exist
  const startDateTimeInput = document.getElementById("start-date-time");
  const endDateTimeInput = document.getElementById("end-date-time");

  if (startDateTimeInput) {
    startDateTimeInput.value = formatDateWithMidnight(startDateTime);
  }

  if (endDateTimeInput) {
    endDateTimeInput.value = formatDate(endDateTime);
  }

  // Try to update filter history text if elements exist
  const filterHistoryStart = document.getElementById("filter_history_start");
  const filterHistoryEnd = document.getElementById("filter_history_end");

  if (filterHistoryStart) {
    filterHistoryStart.textContent = formatDateForDisplay(startDateTime);
  }

  if (filterHistoryEnd) {
    filterHistoryEnd.textContent = formatDateForDisplay(endDateTime);
  }

  startDatetimeFilter = formatToISO(startDateTime);
  endDatetimeFilter = formatToISO(endDateTime);
}

function preventDoubleClick(button, action, delay = 1000) {
  if (button.disabled) return;
  
  button.disabled = true;
  action();
  
  setTimeout(() => {
      button.disabled = false;
  }, delay);
}