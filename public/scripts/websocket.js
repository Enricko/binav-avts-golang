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
  websocket.onclose = () => setTimeout(connectWebSocket, WEBSOCKET_RECONNECT_DELAY);
  websocket.onerror = () => websocket.close();
}

function handleWebSocketMessage(event) {
  const { type, payload } = JSON.parse(event.data);
  const handlers = {
    realtime_update: handleRealtimeVessel,
    vessel_records_start: handleVesselRecordsStart,
    vessel_records_count: handleVesselRecordsCount,
    vessel_records_batch: handleVesselRecordsBatch,
    vessel_records_complete: handleVesselRecordsComplete
  };
  
  if (handlers[type]) {
    handlers[type](payload);
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

async function handleVesselRecordsStart(payload) {
  await clearPolylines();
  vesselHistoryData = [];
  isProcessingComplete = false;
  fetchStartTime = Date.now();
  document.getElementById("spinner").style.display = "block";
  document.getElementById("fetch_time").textContent = "Fetching...";
  vesselHistoryData.kapal = payload.kapal;
  btnLoad.disabled = true;
}

function handleVesselRecordsCount(payload) {
  totalVesselHistoryRecords = payload.total_records;
  document.getElementById("total_records").textContent = totalVesselHistoryRecords;
}

function handleVesselRecordsBatch(records) {
  const newRecords = records.map((record) => ({
    record,
    latlng: {
      lat: convertDMSToDecimal(record.latitude),
      lng: convertDMSToDecimal(record.longitude),
    },
    dateTime: record.created_at,
    status: record.telnet_status,
  }));

  vesselHistoryData.push(...newRecords);

  document.getElementById("fetch_time").textContent = 
    `Received ${vesselHistoryData.length} of ${totalVesselHistoryRecords} records`;

  if (isProcessingComplete && vesselHistoryData.length <= totalVesselHistoryRecords) {
    processCompletedData();
  }
}

function handleVesselRecordsComplete() {
  isProcessingComplete = true;
  
  if (vesselHistoryData.length >= totalVesselHistoryRecords) {
    processCompletedData();
  }
}

async function processCompletedData() {
  const fetchEndTime = Date.now();
  const fetchDuration = fetchEndTime - fetchStartTime;
  displayFetchTime(fetchDuration);

  initializeCompleteHistory();
  await displayVesselHistoryPolyline();
  updateHistoryTable(0);
  // initializeHistoryMarker();
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