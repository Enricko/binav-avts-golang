const websocketUrl = `ws://localhost:8080/ws/kapal`;

let autoCompleteInstance;
let currentDevices = [];
let markers = {}; // Initialize markers object to store marker references
let reconnectInterval = 5000; // 5 seconds

function connectWebSocket() {
  websocket = new WebSocket(websocketUrl);

  websocket.onopen = function () {
    console.log("WebSocket connection established.");
  };

  websocket.onmessage = function (event) {
    const data = JSON.parse(event.data);
    console.log(event.data);
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
    path: "M 16 7 l 0 13 L 8 20 L 8 7 L 12 -1 L 16 7",
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
        path: "M 16 7 l 0 13 L 8 20 L 8 7 L 12 -1 L 16 7",
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

document.getElementById("searchButton").addEventListener("click", function() {
  const searchTerm = document.getElementById("vesselSearch").value.trim();
  console.log("searchTerm");
  if (markers.hasOwnProperty(searchTerm)) {
    const marker = markers[searchTerm];
    smoothPanTo(marker.getPosition());
    map.setZoom(12);
  } else {
    alert("Vessel not found.");
  }
});

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