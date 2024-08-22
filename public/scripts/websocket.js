// Constants
const WEBSOCKET_RECONNECT_DELAY = 5000; // 5 seconds

// Global variables
let websocket;
let isProcessingComplete = false;
let fetchStartTime;

// WebSocket functions
function connectWebSocket() {
  websocket = new WebSocket(websocketUrl);
  websocket.onopen = () => console.log("WebSocket connected");
  websocket.onmessage = handleWebSocketMessage;
  websocket.onclose = () =>
    setTimeout(connectWebSocket, WEBSOCKET_RECONNECT_DELAY);
  websocket.onerror = (error) => {
    console.error("WebSocket error:", error);
    websocket.close();
  };
}

function handleWebSocketMessage(event) {
  const message = JSON.parse(event.data);
  switch (message.type) {
    case "realtime_update":
      handleRealtimeVessel(message.payload);
      break;
    case "vessel_records_start":
      handleVesselRecordsStart(message.payload);
      break;
    case "vessel_records_count":
      handleVesselRecordsCount(message.payload);
      break;
    case "vessel_records_batch":
      handleVesselRecordsBatch(message.payload);
      break;
    case "vessel_records_complete":
      handleVesselRecordsComplete(message.payload);
      break;
    default:
      console.log("Unknown message type:", message.type);
  }
}

function handleRealtimeVessel(data) {
  const newDevices = Object.keys(data);

  const sortedNewDevices = newDevices.sort();
  const sortedCurrentDevices = currentDevices.sort();

  if (sortedNewDevices.toString() !== sortedCurrentDevices.toString()) {
    currentDevices = newDevices;
    updateAutoComplete(currentDevices);
  }

  for (const device in data) {
    dataDevices[device] = data[device];
    updateMarkerIfNeeded(device, data[device]);
  }
}

function handleVesselRecordsStart(payload) {
  vesselHistoryData = [];
  isProcessingComplete = false;
  fetchStartTime = Date.now();
  document.getElementById("spinner").style.display = "block";
  document.getElementById("fetch_time").textContent = "Fetching...";
  vesselHistoryData.kapal = payload.kapal;
  // console.log(ve);
}

function handleVesselRecordsCount(payload) {
  totalVesselHistoryRecords = payload.total_records;
  document.getElementById("total_records").textContent =
    totalVesselHistoryRecords;
}

function handleVesselRecordsBatch(records) {
  const newRecords = records.map((record) => ({
    record: record,
    latlng: {
      lat: convertDMSToDecimal(record.latitude),
      lng: convertDMSToDecimal(record.longitude),
    },
    dateTime: record.created_at,
    status: record.telnet_status,
  }));

  vesselHistoryData.push(...newRecords);

  document.getElementById("fetch_time").textContent = `Received ${vesselHistoryData.length} of ${totalVesselHistoryRecords} records`;

  if (isProcessingComplete && vesselHistoryData.length <= totalVesselHistoryRecords) {
    processCompletedData();
}
}

function handleVesselRecordsComplete(payload) {
  // console.log(`Finished receiving records. Processing time: ${payload.processing_time} ms`);
  isProcessingComplete = true;
  
  // Only process the data if we've received all expected records
  if (vesselHistoryData.length >= totalVesselHistoryRecords) {
    processCompletedData(payload);
  } else {
      console.log("Waiting for final batches...");
  }
}

function processCompletedData(payload) {
  const fetchEndTime = Date.now();
  const fetchDuration = fetchEndTime - fetchStartTime;
  displayFetchTime(fetchDuration);

  initializeCompleteHistory(payload);

  displayVesselHistoryPolyline();
  initializeHistoryMarker();
  updateHistoryTable(0);
}

function displayFetchTime(duration) {
  let displayText;
  if (duration < 1000) {
      displayText = `${duration} ms`;
  } else if (duration < 60000) {
      displayText = `${(duration / 1000).toFixed(2)} sec`;
  } else {
      const minutes = Math.floor(duration / 60000);
      const seconds = ((duration % 60000) / 1000).toFixed(2);
      displayText = `${minutes} min ${seconds} sec`;
  }
  document.getElementById("fetch_time").textContent = displayText;
}