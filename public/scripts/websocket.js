let autoCompleteInstance;
let currentDevices = [];
let markers = {};
let dataDevices = {};
let reconnectInterval = 5000;
let markerLabel = new google.maps.InfoWindow();
let currentSelectedMarker;

function connectWebSocket() {
  websocket = new WebSocket(websocketUrl);

  websocket.onopen = () => console.log("WebSocket connection established.");

  websocket.onmessage = (event) => handleWebSocketMessage(event);

  websocket.onclose = () => {
    console.log("WebSocket connection closed. Reconnecting...");
    setTimeout(connectWebSocket, reconnectInterval);
  };

  websocket.onerror = (error) => {
    console.log("WebSocket error: ", error);
    websocket.close();
  };
}

function handleWebSocketMessage(event) {
  const data = JSON.parse(event.data);
  for (const device in data) {
    if (data.hasOwnProperty(device)) {
      const newDevices = Object.keys(data);
      if (data[device].kapal.status === true) {
        if (newDevices.sort().toString() !== currentDevices.sort().toString()) {
          currentDevices = newDevices;
          updateAutoComplete(currentDevices);
        }

        dataDevices[device] = data[device];

        if (
          data[device].nmea.latitude != "" &&
          data[device].nmea.longitude != ""
        ) {
          updateMarkerPosition(
            device,
            convertDMSToDecimal(data[device].nmea.latitude),
            convertDMSToDecimal(data[device].nmea.longitude),
            data[device].nmea.heading_degree,
            data[device].kapal.width_m,
            data[device].kapal.height_m,
            data[device].kapal.top_range,
            data[device].kapal.left_range,
            data[device].kapal.image_map
          );
        }
      }
    }
  }
}

function parseGGA(gga) {
  const gpsQualitys = [
    "Fix not valid",
    "GPS fix",
    "Differential GPS fix",
    "Not applicable",
    "RTK Fixed",
    "RTK Float",
    "INS Dead reckoning",
  ];
  const fields = gga.split(",");
  const latitudeDMS = parseFloat(fields[2]);
  const latitudeDirection = fields[3];
  const longitudeDMS = parseFloat(fields[4]);
  const longitudeDirection = fields[5];
  const gpsQuality = gpsQualitys[parseInt(fields[6])];

  const latitude = convertDMSToDecimal(latitudeDMS, latitudeDirection);
  const longitude = convertDMSToDecimal(longitudeDMS, longitudeDirection);

  let LatMinute = `${latitudeDMS},${latitudeDirection}`;
  let LongMinute = `${longitudeDMS},${longitudeDirection}`;
  let formattedLatLong = convertCoordinates(LatMinute, LongMinute);

  return {
    latitude,
    latMinute: formattedLatLong.lat,
    longitude,
    longMinute: formattedLatLong.long,
    gpsQuality,
  };
}

function parseHDT(hdt) {
  const fields = hdt.split(",");
  return parseFloat(fields[1]);
}

function parseVTG(vtg) {
  const parts = vtg.split(",");
  const courseTrue = parseFloat(parts[1]);
  const courseMagnetic = parts[3] !== "" ? parseFloat(parts[3]) : null;
  const speedKnots = parseFloat(parts[5]);
  const speedKmh = parseFloat(parts[7]);
  const modeIndicator = parts[9];
  const modeIndicatorText = getModeIndicatorText(modeIndicator);

  return {
    courseTrue,
    courseMagnetic,
    speedKnots,
    speedKmh,
    modeIndicator,
    modeIndicatorText,
  };
}

function getModeIndicatorText(modeIndicator) {
  const modeTexts = {
    A: "Autonomous mode",
    D: "Differential mode",
    E: "Estimated (dead reckoning) mode",
    M: "Manual Input mode",
    S: "Simulator mode",
    N: "Data not valid",
  };
  return modeTexts[modeIndicator] || "Unknown";
}
function convertDMSToDecimal(degreeMinute) {
  // Extract the degree and minute components
  const degreePattern = /^(\d+)°(\d+\.\d+)°([NS|EW])$/;
  const match = degreePattern.exec(degreeMinute.trim());

  if (!match) {
    throw new Error(`Invalid degree-minute format: "${degreeMinute}"`);
  }

  const degrees = parseFloat(match[1]);
  const minutes = parseFloat(match[2]);
  const direction = match[3];

  // Convert to decimal degrees
  let decimalDegrees = degrees + minutes / 60;

  // Adjust for direction
  if (direction === "S" || direction === "W") {
    decimalDegrees = -decimalDegrees;
  }

  return decimalDegrees;
}

function convertCoordinates(latInput, longInput) {
  function formatCoordinate(coordinate) {
    const parts = coordinate.split(",");
    const value = parseFloat(parts[0]);
    const hemisphere = parts[1].trim();
    const degrees = Math.floor(value / 100);
    const minutes = (value % 100).toFixed(4);
    return `${degrees}\u00B0${minutes}\u00B0${hemisphere}`;
  }

  const lat = formatCoordinate(latInput);
  const long = formatCoordinate(longInput);
  return { lat, long };
}

async function updateMarkerPosition(
  device,
  latitude,
  longitude,
  heading,
  width,
  height,
  top,
  left,
  imageMap
) {
  let latMinute = dataDevices[device].nmea.latitude;
  let longMinute = dataDevices[device].nmea.longitude;
  const contentString = createInfoWindowContent(device, latMinute, longMinute);

  if (markers.hasOwnProperty(device)) {
    markers[device].update(
      device,
      { lat: latitude, lng: longitude },
      top,
      left,
      width,
      height,
      (heading +
        dataDevices[device].kapal.calibration +
        dataDevices[device].kapal.heading_direction) %
        360,
      imageMap,
      contentString,
      dataDevices[device].nmea.status
    );

    
  } else {
    markers[device] = new VesselOverlay(
      map,
      device,
      { lat: latitude, lng: longitude },
      top,
      left,
      width,
      height,
      (heading +
        dataDevices[device].kapal.calibration +
        dataDevices[device].kapal.heading_direction) %
        360,
      imageMap,
      contentString,
      dataDevices[device].nmea.status
    );
  }
}

function createInfoWindowContent(device, latitude, longitude) {
  return `
    <div id="content">
      <div id="siteNotice"></div>
      <h1 id="firstHeading" class="firstHeading">${device}</h1>
      <div id="bodyContent">
        <p>Latitude: ${latitude}<br>Longitude: ${longitude}</p>
      </div>
    </div>`;
}



function dataKapalMarker(device) {
  const data = dataDevices[device];
  document.getElementById("heading_hdt").textContent = `${
    data.nmea.heading_degree + data.kapal.calibration
  }\u00B0`;
  document.getElementById(
    "SOG"
  ).textContent = `${data.nmea.speed_in_knots} KTS`;
  document.getElementById("vesselName").textContent = device;
  document.getElementById("status_telnet").textContent = data.nmea.status;
  document.getElementById("status_telnet").style.color =
    data.nmea.status == "Connected" ? "green" : "red";
  document.getElementById("latitude").textContent = data.nmea.latitude;
  document.getElementById("longitude").textContent = data.nmea.longitude;
  document.getElementById("SOLN").textContent = data.nmea.gps_quality_indicator;
  document.getElementById("water_depth").textContent = "DBT " + data.nmea.water_depth;
}



function getDataKapalMarker(device) {
  if (vesselPolylineHistory) vesselPolylineHistory.setMap(null);
  if (markerStrava) markerStrava.setMap(null);
  btnPlay.disabled = true; 

  const vessel_record_preview = document.getElementById("vessel_record_preview");
  dataKapalMarker(device);
  currentSelectedMarker = device;
  vessel_record_preview.style.display = "block";
  isPreview = true;
  viewDetailKapal();
  if(isPlayingAnimation){
    dismissHistory();
  }
}