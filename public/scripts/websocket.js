const websocketUrl = `ws://localhost:8080/ws/kapal`;

let autoCompleteInstance;
let currentDevices = [];
let markers = {}; // Initialize markers object to store marker references
let dataDevices = {};
let reconnectInterval = 5000; // 5 seconds
let markerLabel = new google.maps.InfoWindow(); // Define a single global markerLabel
let currentSelectedMarker;

function connectWebSocket() {
  websocket = new WebSocket(websocketUrl);

  websocket.onopen = function () {
    console.log("WebSocket connection established.");
  };

  websocket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    for (const device in data) {
      if (data.hasOwnProperty(device)) {
        const newDevices = Object.keys(data);

        // Check if there is a change in device data
        if (newDevices.sort().toString() !== currentDevices.sort().toString()) {
          currentDevices = newDevices;
          updateAutoComplete(currentDevices);
        }
        dataDevices[device] = data[device];

        const gga = data[device].nmea.gga;
        const hdt = data[device].nmea.hdt;
        const vtg = data[device].nmea.vtg;
        const parsedGGA = parseGGA(gga);
        const heading = parseHDT(hdt);
        
        updateMarkerPosition(
          device,
          parsedGGA.latitude,
          parsedGGA.longitude,
          heading
        );
      }
    }
  };

  websocket.onclose = function (event) {
    console.log("WebSocket connection closed. Reconnecting...");
    setTimeout(connectWebSocket, reconnectInterval);
  };

  websocket.onerror = function (error) {
    console.log("WebSocket error: ", error);
    websocket.close();
  };
}

// Parse NMEA GGA sentence
function parseGGA(gga) {
  const gpsQualitys = [
    "Fix not valid",
    "GPS fix",
    "Differential GPS fix (DGNSS), SBAS, OmniSTAR VBS, Beacon, RTX in GVBS mode",
    "Not applicable",
    "RTK Fixed, xFill",
    "RTK Float, OmniSTAR XP/HP",
    "Location RTK, RTX",
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

function convertCoordinates(latInput, longInput) {
  // Function to parse and format the latitude or longitude
  function formatCoordinate(coordinate) {
    const parts = coordinate.split(",");
    const value = parseFloat(parts[0]);
    const hemisphere = parts[1].trim();

    // Extract degrees and minutes
    const degrees = Math.floor(value / 100);
    const minutes = (value % 100).toFixed(4);

    // Format the coordinate string
    return `${degrees}\u00B0${minutes}\u00B0${hemisphere}`;
  }

  // Format latitude and longitude
  const lat = formatCoordinate(latInput);
  const long = formatCoordinate(longInput);

  // Return an object with lat and long properties
  return { lat, long };
}

// Parse NMEA HDT sentence
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
  let modeIndicatorText = "";
  switch (modeIndicator) {
    case "A":
      modeIndicatorText = "Autonomous mode";
      break;
    case "D":
      modeIndicatorText = "Differential mode";
      break;
    case "E":
      modeIndicatorText = "Estimated (dead reckoning) mode";
      break;
    case "M":
      modeIndicatorText = "Manual Input mode";
      break;
    case "S":
      modeIndicatorText = "Simulator mode";
      break;
    case "N":
      modeIndicatorText = "Data not valid";
      break;
    default:
      modeIndicatorText = "Unknown";
  }

  return {
    courseTrue: courseTrue,
    courseMagnetic: courseMagnetic,
    speedKnots: speedKnots,
    speedKmh: speedKmh,
    modeIndicator: modeIndicator,
  };
}

function convertDMSToDecimal(degrees, direction) {
  // Extract degrees and minutes
  const d = Math.floor(degrees / 100);
  const m = degrees % 100;

  // Convert to decimal degrees
  let decimalDegrees = d + m / 60;

  // Adjust for negative direction
  if (direction === "S" || direction === "W") {
    decimalDegrees = -decimalDegrees;
  }

  return decimalDegrees;
}

// Update marker position or create new marker
async function updateMarkerPosition(device, latitude, longitude, heading) {
  const boatIcon = {
    path: "M 14 30 L 2 30 L 2 -10 L 14 -10 L 14 30 V -6 M 12 -9 C 12 -9 11 -9 11 -8 C 11 -7 12 -7 12 -7 C 12 -7 13 -7 13 -8 C 13 -8 13 -9 12 -9 M 3 29 L 13 29 L 13 14 L 3 14 L 3 29",
    fillColor: "#ffd400",
    fillOpacity: 1,
    strokeColor: "#000",
    strokeOpacity: 0.4,
    scale: calculateMarkerSize(map.getZoom()),
    rotation: (heading + dataDevices[device].kapal.heading_direction) % 360,
    anchor: new google.maps.Point(13, 13),
  };
  const ggaKapal = parseGGA(dataDevices[device].nmea.gga);

  if (!markers.hasOwnProperty(device)) {
    markers[device] = new google.maps.Marker({
      position: { lat: latitude, lng: longitude },
      map: map,
      title: device,
      icon: boatIcon,
    });


    markers[device].addListener("dblclick", function () {
      getDataKapalMarker(device);
    });

    // Add hover event listener
    markers[device].addListener("mouseover", function () {
      updateInfoWindow(device, ggaKapal.latMinute, ggaKapal.longMinute, markers[device]);
    });

    markers[device].addListener("mouseout", function () {
      markerLabel.close();
    });
  } else {
    markers[device].setPosition({ lat: latitude, lng: longitude });
    markers[device].setIcon(boatIcon);

    // Ensure the info window content is updated
    google.maps.event.clearListeners(markers[device], "mouseover"); // Clear the previous 'mouseover' listener
    markers[device].addListener("mouseover", function () {
      updateInfoWindow(device, ggaKapal.latMinute, ggaKapal.longMinute, markers[device]);
    });

    markers[device].addListener("mouseout", function () {
      markerLabel.close();
    });
  }
}

function updateInfoWindow(device, latitude, longitude, marker) {
  const contentString = `
    <div id="content">
      <div id="siteNotice"></div>
      <h1 id="firstHeading" class="firstHeading">${device}</h1>
      <div id="bodyContent">
        <p>Latitude: ${latitude}<br>Longitude: ${longitude}</p>
      </div>
    </div>`;

  markerLabel.setContent(contentString);
  markerLabel.open(map, marker);
}

// Update marker sizes based on zoom level
function updateMarkerSizes(zoom) {
  for (const device in markers) {
    if (markers.hasOwnProperty(device)) {
      const marker = markers[device];
      marker.setIcon({
        path: "M 14 30 L 2 30 L 2 -10 L 14 -10 L 14 30 V -6 M 12 -9 C 12 -9 11 -9 11 -8 C 11 -7 12 -7 12 -7 C 12 -7 13 -7 13 -8 C 13 -8 13 -9 12 -9 M 3 29 L 13 29 L 13 14 L 3 14 L 3 29",
        fillColor: "#ffd400",
        fillOpacity: 1,
        strokeColor: "#000",
        strokeOpacity: 0.4,
        scale: calculateMarkerSize(zoom),
        rotation: marker.getIcon().rotation,
        anchor: new google.maps.Point(13, 13),
      });
    }
  }
}

function smoothPanTo(latLng) {
  const panSteps = 30; // Number of steps for the pan animation
  const panDuration = 1000; // Duration of the pan animation in milliseconds
  const panInterval = panDuration / panSteps; // Interval between each step

  const startLat = map.getCenter().lat();
  const startLng = map.getCenter().lng();
  const endLat = latLng.lat();
  const endLng = latLng.lng();

  let step = 0;

  function panStep() {
    const lat = startLat + (endLat - startLat) * (step / panSteps);
    const lng = startLng + (endLng - startLng) * (step / panSteps);
    map.setCenter({ lat: lat, lng: lng });

    if (step < panSteps) {
      step++;
      setTimeout(panStep, panInterval);
    }
  }

  panStep();
}

function getDataKapalMarker(device) {
  switchWindow(true);
  dataKapalMarker(device);
  currentSelectedMarker = device;
}

function dataKapalMarker(device) {
  let data = dataDevices[device];
  let ggaData = parseGGA(data.nmea.gga);
  let hdtData = parseHDT(data.nmea.hdt);
  let vtgData = parseVTG(data.nmea.vtg);
  document.getElementById("vesselName").textContent = device;
  document.getElementById("status_telnet").textContent = data.nmea.status;
  document.getElementById("status_telnet").style.color = data.nmea.status ? "green" : "red";
  document.getElementById("latitude").textContent = ggaData.latMinute;
  document.getElementById("longitude").textContent = ggaData.longMinute;
  document.getElementById("true_heading").textContent = `${(hdtData + data.kapal.heading_direction) % 360}\u00B0`;
  document.getElementById("heading_hdt").textContent = hdtData+"\u00B0";
  document.getElementById("calibration").textContent = data.kapal.heading_direction + "\u00B0";
  document.getElementById("SOG").textContent = vtgData.speedKnots + " KTS";
  document.getElementById("SOLN").textContent = ggaData.gpsQuality;
}

function switchWindow(onoff) {
  // var windowButton = document.getElementById("detail-window");
  var hideButton = document.getElementById("hideButton");
  if (onoff) {
    // windowButton.classList.remove("d-none");
    hideButton.classList.remove("d-none");
  } else {
    // windowButton.classList.add = "d-none";
    hideButton.classList.add = "d-none";
  }
}
var timeoutID;
document.getElementById("hideButton").addEventListener("click", function () {
  var container = document.getElementById("myContainer");
  if (container.style.display === "none") {
    container.style.display = "block";
    timeoutID = setInterval(function () {
      dataKapalMarker(currentSelectedMarker);
    }, 500);
  } else {
    container.style.display = "none";
    clearInterval(timeoutID);
  }
});
