const websocketUrl = `ws://localhost:8080/ws/kapal`;

let autoCompleteInstance;
let currentDevices = [];
let markers = {}; // Initialize markers object to store marker references

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
        const gga = data[device].nmea.gga;
        const hdt = data[device].nmea.hdt;
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
  const fields = gga.split(",");
  const latitudeDMS = parseFloat(fields[2]);
  const latitudeDirection = fields[3];
  const longitudeDMS = parseFloat(fields[4]);
  const longitudeDirection = fields[5];

  const latitude = convertDMSToDecimal(latitudeDMS, latitudeDirection);
  const longitude = convertDMSToDecimal(longitudeDMS, longitudeDirection);

  // console.log('Latitude:', latitude, 'Longitude:', longitude);
  return { latitude, longitude };
}

// Parse NMEA HDT sentence
function parseHDT(hdt) {
  const fields = hdt.split(",");
  return parseFloat(fields[1]);
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
    path: "M14 8.947L22 14v2l-8-2.526v5.36l3 1.666V22l-4.5-1L8 22v-1.5l3-1.667v-5.36L3 16v-2l8-5.053V3.5a1.5 1.5 0 0 1 3 0v5.447z",
    fillColor: "#ffd400",
    fillOpacity: 1,
    strokeColor: "#000",
    strokeOpacity: 0.4,
    scale: calculateMarkerSize(map.getZoom()),
    rotation: heading,
    anchor: new google.maps.Point(13, 13),
  };

  if (!markers.hasOwnProperty(device)) {
    markers[device] = new google.maps.Marker({
      position: { lat: latitude, lng: longitude },
      map: map,
      title: device,
      icon: boatIcon,
    });
  } else {
    markers[device].setPosition({ lat: latitude, lng: longitude });
    markers[device].setIcon(boatIcon); // Update marker icon
  }
}

// Update marker sizes based on zoom level
function updateMarkerSizes(zoom) { 
  for (const device in markers) {
    if (markers.hasOwnProperty(device)) {
      const marker = markers[device];
      marker.setIcon({
        path: "M14 8.947L22 14v2l-8-2.526v5.36l3 1.666V22l-4.5-1L8 22v-1.5l3-1.667v-5.36L3 16v-2l8-5.053V3.5a1.5 1.5 0 0 1 3 0v5.447z",
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
