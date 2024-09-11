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
    timeZone: "Asia/Jakarta",
  };

  const formattedDate = date.toLocaleString("en-GB", options).replace(",", "");
  return formattedDate.replace(/\//g, "-");
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
    timeZone: "Asia/Jakarta",
  };

  const formattedDate = date.toLocaleString("en-GB", options).replace(",", "");
  return formattedDate;
}

const formatDate = (date) => {
  const pad = (num) => String(num).padStart(2, "0");
  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  const seconds = pad(date.getSeconds());
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
};

// Format the date for the start time with 00:00:00
const formatDateWithMidnight = (date) => {
  const pad = (num) => String(num).padStart(2, "0");
  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  return `${year}-${month}-${day} 00:00:00`;
};

function startToEndDatetimeFilterForm() {
  // Get current date and time
  const now = new Date();

  // Set default end date to now
  const endDateTime = now;

  // Set default start date to 3 days ago with time 00:00:00
  const startDateTime = new Date(now);
  startDateTime.setDate(startDateTime.getDate() - 3);

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
    filterHistoryStart.textContent = formatDateWithMidnight(startDateTime);
  }

  if (filterHistoryEnd) {
    filterHistoryEnd.textContent = formatDate(endDateTime);
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