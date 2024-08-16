let currentDevices = [];
let markers = {};
let dataDevices = {};
let websocket;

function connectWebSocket() {
  websocket = new WebSocket(websocketUrl);
  websocket.onopen = () => console.log("WebSocket connected");
  websocket.onmessage = handleWebSocketMessage;
  websocket.onclose = () => setTimeout(connectWebSocket, 5000);
  websocket.onerror = (error) => {
    console.error("WebSocket error:", error);
    websocket.close();
  };
}

function handleWebSocketMessage(event) {
  const data = JSON.parse(event.data);
  const newDevices = Object.keys(data);

  // Sort both arrays to ensure consistent comparison
  const sortedNewDevices = newDevices.sort();
  const sortedCurrentDevices = currentDevices.sort();

  // Check if the devices have changed
  if (sortedNewDevices.toString() !== sortedCurrentDevices.toString()) {
    currentDevices = newDevices; // Update currentDevices
    updateAutoComplete(currentDevices); // Call updateAutoComplete with the new devices
  }

  for (const device in data) {
    // if (data[device].kapal.status) {
      dataDevices[device] = data[device];
      updateMarkerIfNeeded(device, data[device]);
    // }
  }
}

function updateMarkerIfNeeded(device, data) {
  const { nmea, kapal } = data;
  if (nmea.latitude && nmea.longitude) {
    const position = {
      lat: convertDMSToDecimal(nmea.latitude),
      lng: convertDMSToDecimal(nmea.longitude)
    };
    const heading = (nmea.heading_degree + kapal.calibration + kapal.heading_direction) % 360;
    const contentString = createInfoWindowContent(device, nmea.latitude, nmea.longitude);

    if (markers[device]) {
      markers[device].update(device, position, kapal.top_range, kapal.left_range, kapal.width_m, kapal.height_m, heading, kapal.image_map, contentString, nmea.status);
    } else {
      markers[device] = new VesselOverlay(map, device, position, kapal.top_range, kapal.left_range, kapal.width_m, kapal.height_m, heading, kapal.image_map, contentString, nmea.status);
    }
  }
}

function convertDMSToDecimal(degreeMinute) {
  const [degrees, minutes, direction] = degreeMinute.match(/(\d+)°(\d+\.\d+)°([NS|EW])/).slice(1);
  let decimalDegrees = parseFloat(degrees) + parseFloat(minutes) / 60;
  return direction === "S" || direction === "W" ? -decimalDegrees : decimalDegrees;
}

function createInfoWindowContent(device, latitude, longitude) {
  return `
    <div id="content">
      <h1 id="firstHeading" class="firstHeading">${device}</h1>
      <div id="bodyContent">
        <p>Latitude: ${latitude}<br>Longitude: ${longitude}</p>
      </div>
    </div>`;
}

function dataKapalMarker(device) {
  const data = dataDevices[device];
  if (!data) return;

  const elements = {
    heading_hdt: `${data.nmea.heading_degree + data.kapal.calibration}°`,
    SOG: `${data.nmea.speed_in_knots} KTS`,
    vesselName: device,
    status_telnet: data.nmea.status,
    latitude: data.nmea.latitude,
    longitude: data.nmea.longitude,
    SOLN: data.nmea.gps_quality_indicator,
    water_depth: `${formatWaterDepthNumber(data.nmea.water_depth)} Meter`
  };

  for (const [id, value] of Object.entries(elements)) {
    const element = document.getElementById(id);
    if (element) {
      if (id === 'status_telnet') {
        element.textContent = value;
        element.style.color = value === "Connected" ? "green" : "red";
      } else {
        element.textContent = value;
      }
    }
  }
}

function formatWaterDepthNumber(number) {
  const [part1, part2] = number.toString().padStart(3, '0').match(/^(\d+)(\d{2})$/).slice(1);
  return parseFloat(`${part1 || '0'}.${part2}`);
}

function getDataKapalMarker(device) {
  const vessel_record_preview = document.getElementById("vessel_record_preview");
  dataKapalMarker(device);
  startToEndDatetimeFilterForm();

  if (currentSelectedMarker !== device) {
    currentSelectedMarker = device;
    resetVesselState();
    vessel_record_preview.style.display = "block";
    isPreviewActive = true;
    toggleVesselDetailSidebar();
    defaultHistoryTable();
    document.getElementById("total_records").textContent = "0";
  }
}

function resetVesselState() {
  btnPlay.disabled = btnDownloadCSV.disabled = true;
  vesselPolylineHistory = [];
  if (historyMarker) {
    historyMarker.setMap(null);
    historyMarker = null;
  }
  resetVesselHistoryAnimation();
}

function resetVesselHistoryAnimation() {
  progressSlider.value = progressSlider.max = totalVesselHistoryRecords = currentAnimationIndex = 0;
  if (isAnimationPlaying) stopVesselHistoryAnimation();
}

// Initialize WebSocket connection
connectWebSocket();