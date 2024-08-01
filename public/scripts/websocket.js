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
      contentString
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
      contentString
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
}


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

    // Determine the color based on device status
    const status = dataDevices[this.device].nmea.status;
    const color = status === "Connected" ? [40, 167, 69] : [220, 53, 69]; // Green if Connected, Red if Disconnected

    // Use the utility function to change the image color and add shadow
    changeImageColor(
      `/public/assets/images/${this.imageMap}`,
      color,
      (dataUrl) => {
        if (dataUrl) {
          img.src = dataUrl;
        } else {
          img.src = `/public/assets/images/${this.imageMap}`; // Fallback if processing fails
        }
      }
    );
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

      // Determine the color based on device status
      const status = dataDevices[this.device].nmea.status;
      const color = status === "Connected" ? [40, 167, 69] : [220, 53, 69]; // Green if Connected, Red if Disconnected

      // Use the utility function to change the image color and add shadow
      changeImageColor(
        `/public/assets/images/${this.imageMap}`,
        color,
        (dataUrl) => {
          if (dataUrl) {
            img.src = dataUrl;
          } else {
            img.src = `/public/assets/images/${this.imageMap}`; // Fallback if processing fails
          }
        }
      );

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

function changeImageColor(imageUrl, color, callback) {
  const img = new Image();
  img.crossOrigin = "Anonymous";
  img.src = imageUrl;

  img.onload = function () {
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");

    canvas.width = img.width;
    canvas.height = img.height;

    // Draw the image to get the original shape
    ctx.drawImage(img, 0, 0);

    // Save the current context state
    ctx.save();

    // Set shadow properties
    ctx.shadowColor = "rgba(0, 0, 0, 0.5)"; // Shadow color
    ctx.shadowBlur = 10; // Shadow blur
    ctx.shadowOffsetX = 0; // Shadow offset X
    ctx.shadowOffsetY = 0; // Shadow offset Y

    // Draw the image again to apply the shadow
    ctx.drawImage(img, 0, 0);

    // Restore the context to remove shadow effect for further drawing
    ctx.restore();

    // Get the image data to change colors
    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
    const data = imageData.data;

    const [r, g, b] = color;

    for (let i = 0; i < data.length; i += 4) {
      if (data[i + 3] !== 0) {
        // Check if the pixel is not transparent
        data[i] = r; // Red
        data[i + 1] = g; // Green
        data[i + 2] = b; // Blue
        // data[i + 3] = data[i + 3]; // Alpha (unchanged)
      }
    }

    ctx.putImageData(imageData, 0, 0);
    callback(canvas.toDataURL());
  };

  img.onerror = function () {
    console.error("Failed to load image:", imageUrl);
    callback(null);
  };
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
    dismissStrava();
  }
}