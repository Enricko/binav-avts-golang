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
        if (
          data[device].kapal.status == true 
          // &&
          // data[device].nmea.gga != "No Data"
        ) {
          // Check if there is a change in device data
          if (
            newDevices.sort().toString() !== currentDevices.sort().toString()
          ) {
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
            heading,
            data[device].kapal.width_m,
            data[device].kapal.height_m,
            data[device].kapal.top_range,
            data[device].kapal.left_range,
            data[device].kapal.image_map
          );
        }
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
    "OmniSTAR VBS",
    "Not applicable",
    "RTK Fixed, xFill",
    "OmniSTAR XP/HP",
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
  const ggaKapal = parseGGA(dataDevices[device].nmea.gga);
  if (markers.hasOwnProperty(device)) {
    // Update existing overlay
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
      `<div id="content">
      <div id="siteNotice"></div>
      <h1 id="firstHeading" class="firstHeading">${device}</h1>
      <div id="bodyContent">
        <h6>Latitude: ${ggaKapal.latMinute}<br>Longitude: ${ggaKapal.longMinute}</h6>
      </div>
    </div>`
    );
  } else {
    // Create new overlay
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
      `<div id="content">
      <div id="siteNotice"></div>
      <h1 id="firstHeading" class="firstHeading">${device}</h1>
      <div id="bodyContent">
        <p>Latitude: ${ggaKapal.latMinute}<br>Longitude: ${ggaKapal.longMinute}</p>
      </div>
    </div>`
    );
  }
  // const boatIcon = {
  //   path: "M 14 30 L 2 30 L 2 -10 L 14 -10 L 14 30 V -6 M 12 -9 C 12 -9 11 -9 11 -8 C 11 -7 12 -7 12 -7 C 12 -7 13 -7 13 -8 C 13 -8 13 -9 12 -9 M 3 29 L 13 29 L 13 14 L 3 14 L 3 29",
  //   fillColor: "#ffd400",
  //   fillOpacity: 1,
  //   strokeColor: "#000",
  //   strokeOpacity: 0.4,
  //   scale: calculateMarkerSize(map.getZoom()),
  //   rotation: (heading + dataDevices[device].kapal.heading_direction) % 360,
  //   anchor: new google.maps.Point(13, 13),
  // };

  // if (!markers.hasOwnProperty(device)) {
  //   markers[device] = new google.maps.Marker({
  //     position: { lat: latitude, lng: longitude },
  //     map: map,
  //     title: device,
  //     icon: boatIcon,
  //   });

  //   markers[device].addListener("dblclick", function () {
  //     getDataKapalMarker(device);
  //   });

  //   // Add hover event listener
  //   markers[device].addListener("mouseover", function () {
  //     updateInfoWindow(device, ggaKapal.latMinute, ggaKapal.longMinute, markers[device]);
  //   });

  //   markers[device].addListener("mouseout", function () {
  //     markerLabel.close();
  //   });
  // } else {
  //   markers[device].setPosition({ lat: latitude, lng: longitude });
  //   markers[device].setIcon(boatIcon);

  //   // Ensure the info window content is updated
  //   google.maps.event.clearListeners(markers[device], "mouseover"); // Clear the previous 'mouseover' listener
  //   markers[device].addListener("mouseover", function () {
  //     updateInfoWindow(device, ggaKapal.latMinute, ggaKapal.longMinute, markers[device]);
  //   });

  //   markers[device].addListener("mouseout", function () {
  //     markerLabel.close();
  //   });
  // }
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

  return contentString;
}

// Update marker sizes based on zoom level
// function updateMarkerSizes(zoom) {
//   for (const device in markers) {
//     if (markers.hasOwnProperty(device)) {
//       const marker = markers[device];
//       marker.setIcon({
//         path: "M 14 30 L 2 30 L 2 -10 L 14 -10 L 14 30 V -6 M 12 -9 C 12 -9 11 -9 11 -8 C 11 -7 12 -7 12 -7 C 12 -7 13 -7 13 -8 C 13 -8 13 -9 12 -9 M 3 29 L 13 29 L 13 14 L 3 14 L 3 29",
//         fillColor: "#ffd400",
//         fillOpacity: 1,
//         strokeColor: "#000",
//         strokeOpacity: 0.4,
//         scale: calculateMarkerSize(zoom),
//         rotation: marker.getIcon().rotation,
//         anchor: new google.maps.Point(13, 13),
//       });
//     }
//   }
// }


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
  document.getElementById("status_telnet").style.color = data.nmea.status == "Connected"
    ? "green"
    : "red";
  document.getElementById("latitude").textContent = ggaData.latMinute;
  document.getElementById("longitude").textContent = ggaData.longMinute;
  document.getElementById("heading_hdt").textContent = hdtData +
  dataDevices[device].kapal.calibration + "\u00B0";
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

class VesselOverlay extends google.maps.OverlayView {
  constructor(
    map,
    device,
    position,
    top,
    left,
    width,
    height,
    rotationAngle,
    imageMap,
    infoContent // New parameter for info window content
  ) {
    super();
    this.map = map;
    this.device = device;
    this.position = position;
    this.vesselWidthMeters = width;
    this.vesselHeightMeters = height;
    this.offsetFromCenter = { x: left, y: top };
    this.rotationAngle = rotationAngle;
    this.imageMap = imageMap;
    this.infoContent = infoContent; // Store the info window content

    this.infoWindow = null; // Info window element
    this.setMap(map);
    // Disable map interactions while overlay is active
    this.disableMapInteractions();
  }

  onAdd() {
    this.div = document.createElement("div");
    this.div.style.borderStyle = "none";
    this.div.style.borderWidth = "0px";
    this.div.style.position = "absolute";
    this.div.style.cursor = "default"; // Ensure the cursor is initially set to default

    const img = document.createElement("img");
    img.src = `/public/assets/images/${this.imageMap}`;
    img.style.width = "100%";
    img.style.height = "100%";
    img.style.pointerEvents = "auto"; // Ensure the image is interactive

    const transformOriginX =
      (this.offsetFromCenter.x / this.vesselWidthMeters) * 100;
    const transformOriginY =
      (this.offsetFromCenter.y / this.vesselHeightMeters) * 100;
    img.style.transformOrigin = `${transformOriginX}% ${transformOriginY}%`;
    img.style.transform = `rotate(${this.rotationAngle}deg)`;

    this.div.appendChild(img);

    const panes = this.getPanes();
    panes.overlayMouseTarget.appendChild(this.div); // Add the div to the overlayMouseTarget pane

    this.div.addEventListener("mouseover", this.onMouseOver.bind(this));
    this.div.addEventListener("mouseout", this.onMouseOut.bind(this));
    this.div.addEventListener("dblclick", this.onDblClick.bind(this));

    // Debugging output
  }

  onDblClick(event) {
    getDataKapalMarker(this.device);
    this.centerMapOnVessel(event);
  }

  onMouseOver() {
    // Handle mouse over event
    this.div.style.cursor = "pointer"; // Change cursor to pointer on hover
    this.showInfoWindow(); // Show the info window on hover
    this.disableMapInteractions();
  }

  onMouseOut() {
    // Handle mouse out event
    this.div.style.cursor = "default"; // Reset cursor to default when not hovering
    this.hideInfoWindow(); // Hide the info window when not hovering
    this.enableMapInteractions();
  }

  draw() {
    const overlayProjection = this.getProjection();
    const positionPixel = overlayProjection.fromLatLngToDivPixel(this.position);

    if (this.div) {
      const scale = this.getScale();
      const scaledWidth = this.vesselWidthMeters * scale;
      const scaledHeight = this.vesselHeightMeters * scale;

      const offsetXPixels = this.metersToPixels(
        this.offsetFromCenter.x,
        this.position.lat,
        scale
      );
      const offsetYPixels = this.metersToPixels(
        this.offsetFromCenter.y,
        this.position.lat,
        scale
      );

      this.div.style.left = positionPixel.x - offsetXPixels + "px";
      this.div.style.top = positionPixel.y - offsetYPixels + "px";
      this.div.style.width = scaledWidth + "px";
      this.div.style.height = scaledHeight + "px";
      this.div.style.zIndex = 999; // Ensure the vessel overlay is on top
    }
  }

  getScale() {
    const zoom = this.map.getZoom();
    const metersPerPixel =
      (156543.03392 * Math.cos((this.position.lat * Math.PI) / 180)) /
      Math.pow(2, zoom);
    return 1 / metersPerPixel;
  }

  metersToPixels(meters, latitude, scale) {
    const metersPerPixel =
      (156543.03392 * Math.cos((latitude * Math.PI) / 180)) /
      Math.pow(2, this.map.getZoom());
    return meters / metersPerPixel;
  }

  update(
    device,
    position,
    top,
    left,
    width,
    height,
    rotationAngle,
    imageMap,
    infoContent
  ) {
    this.position = position;
    this.vesselWidthMeters = width;
    this.vesselHeightMeters = height;
    this.offsetFromCenter = { x: left, y: top };
    this.rotationAngle = rotationAngle;
    this.imageMap = imageMap;
    this.infoContent = infoContent;

    if (this.div) {
      const img = this.div.firstChild;
      img.src = `/public/assets/images/${this.imageMap}`;

      const transformOriginX =
        (this.offsetFromCenter.x / this.vesselWidthMeters) * 100;
      const transformOriginY =
        (this.offsetFromCenter.y / this.vesselHeightMeters) * 100;
      img.style.transformOrigin = `${transformOriginX}% ${transformOriginY}%`;
      img.style.transform = `rotate(${this.rotationAngle}deg)`;

      const overlayProjection = this.getProjection();
      const positionPixel = overlayProjection.fromLatLngToDivPixel(
        this.position
      );

      const scale = this.getScale();
      const scaledWidth = this.vesselWidthMeters * scale;
      const scaledHeight = this.vesselHeightMeters * scale;

      const offsetXPixels = this.metersToPixels(
        this.offsetFromCenter.x,
        this.position.lat,
        scale
      );
      const offsetYPixels = this.metersToPixels(
        this.offsetFromCenter.y,
        this.position.lat,
        scale
      );

      this.div.style.left = positionPixel.x - offsetXPixels + "px";
      this.div.style.top = positionPixel.y - offsetYPixels + "px";
      this.div.style.width = scaledWidth + "px";
      this.div.style.height = scaledHeight + "px";
    } else {
      this.setMap(this.map);
    }
  }

  onRemove() {
    if (this.div) {
      this.div.parentNode.removeChild(this.div);
      this.div = null;
    }
  }

  showInfoWindow() {
    if (!this.infoWindow) {
      this.infoWindow = document.createElement("div");
      this.infoWindow.style.position = "absolute";
      this.infoWindow.style.backgroundColor = "white";
      this.infoWindow.style.border = "1px solid black";
      this.infoWindow.style.padding = "5px";
      this.infoWindow.style.zIndex = 1000; // Ensure the info window is above other elements
      this.infoWindow.innerHTML = this.infoContent;

      this.getPanes().floatPane.appendChild(this.infoWindow);
    }

    const overlayProjection = this.getProjection();
    const positionPixel = overlayProjection.fromLatLngToDivPixel(this.position);

    const scale = this.getScale();
    const offsetXPixels = this.metersToPixels(
      this.offsetFromCenter.x,
      this.position.lat,
      scale
    );
    const offsetYPixels = this.metersToPixels(
      this.offsetFromCenter.y,
      this.position.lat,
      scale
    );

    this.infoWindow.style.left = positionPixel.x - offsetXPixels + "px";
    this.infoWindow.style.top = positionPixel.y - offsetYPixels - 50 + "px"; // Adjust the position as needed
  }

  hideInfoWindow() {
    if (this.infoWindow) {
      this.infoWindow.parentNode.removeChild(this.infoWindow);
      this.infoWindow = null;
    }
  }

  disableMapInteractions() {
    // Disable specific map interactions
    this.map.setOptions({
      disableDoubleClickZoom: true, // Disable zooming on double click
      clickableIcons: false, // Disable clicking on map icons (markers)
    });
  }

  enableMapInteractions() {
    // Enable specific map interactions
    this.map.setOptions({
      disableDoubleClickZoom: false,
      clickableIcons: true,
    });
  }

  centerMapOnVessel(event) {
    // Prevent default double-click behavior (zooming)
    event.preventDefault();

    // Center the map's camera on the vessel
    this.map.setCenter(this.position);
  }
}
