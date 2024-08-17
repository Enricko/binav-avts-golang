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
      lng: convertDMSToDecimal(nmea.longitude),
    };
    const heading =
      (nmea.heading_degree + kapal.calibration + kapal.heading_direction) % 360;
    const contentString = createInfoWindowContent(
      device,
      nmea.latitude,
      nmea.longitude
    );

    if (markers[device]) {
      markers[device].update(
        device,
        position,
        kapal.top_range,
        kapal.left_range,
        kapal.width_m,
        kapal.height_m,
        heading,
        kapal.image_map,
        contentString,
        nmea.status
      );
    } else {
      markers[device] = new VesselOverlay(
        map,
        device,
        position,
        kapal.top_range,
        kapal.left_range,
        kapal.width_m,
        kapal.height_m,
        heading,
        kapal.image_map,
        contentString,
        nmea.status
      );
    }
  }
}

function convertDMSToDecimal(degreeMinute) {
  const [degrees, minutes, direction] = degreeMinute
    .match(/(\d+)°(\d+\.\d+)°([NS|EW])/)
    .slice(1);
  let decimalDegrees = parseFloat(degrees) + parseFloat(minutes) / 60;
  return direction === "S" || direction === "W"
    ? -decimalDegrees
    : decimalDegrees;
}

function createInfoWindowContent(device, latitude, longitude) {
  const data = dataDevices[device];
  if (!data) return "";

  return `
    <div id="content" style="font-family: Arial, sans-serif; width: 300px; padding: 10px;">
      <div style="display: flex; align-items: center; margin-bottom: 15px;">
        <h3 style="margin: 0; flex-grow: 1; font-size: 18px; color: #333;">${device}</h3>
        <img src="/public/upload/assets/image/vessel/${
          data.kapal.image
        }" alt="${device} image" style="width: 80px; height: 80px; object-fit: cover; border-radius: 5px; margin-left: 10px;"/>
      </div>
      <div style="background-color: #f0f0f0; border-radius: 5px; padding: 10px;">
        <p style="margin: 0; line-height: 1.6;">
          <span style="display: inline-block; width: 100px; font-weight: bold;">Latitude:</span> ${latitude}<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">Longitude:</span> ${longitude}<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">Heading:</span> ${
            (data.nmea.heading_degree +
              data.kapal.calibration +
              data.kapal.heading_direction) %
            360
          }°<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">Speed:</span> ${
            data.nmea.speed_in_knots
          } knots<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">Water Depth:</span> ${formatWaterDepthNumber(
            data.nmea.water_depth
          )} meters<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">GPS Quality:</span> ${
            data.nmea.gps_quality_indicator
          }<br>
          <span style="display: inline-block; width: 100px; font-weight: bold;">Status:</span> <span style="color: ${
            data.nmea.status === "Connected" ? "green" : "red"
          };">${data.nmea.status}</span>
        </p>
      </div>
    </div>`;
}

function dataKapalMarker(device) {
  const data = dataDevices[device];
  if (!data) return;

  const elements = {
    heading_hdt_current: `${
      data.nmea.heading_degree + data.kapal.calibration
    }°`,
    SOG_current: `${data.nmea.speed_in_knots} KTS`,
    status_telnet_current: data.nmea.status,
    latitude_current: data.nmea.latitude,
    longitude_current: data.nmea.longitude,
    SOLN_current: data.nmea.gps_quality_indicator,
    water_depth_current: `${formatWaterDepthNumber(
      data.nmea.water_depth
    )} Meter`,
    // Add kapal data
    call_sign_general: data.kapal.call_sign,
    flag_general: data.kapal.flag,
    kelas_general: data.kapal.kelas,
    builder_general: data.kapal.builder,
    year_built_general: data.kapal.year_built,
    heading_direction: `${data.kapal.heading_direction}°`,
    calibration: `${data.kapal.calibration}°`,
    width_m: `${data.kapal.width_m} Meter`,
    height_m: `${data.kapal.height_m} Meter`,
    top_range: data.kapal.top_range,
    left_range: data.kapal.left_range,
    minimum_knot_per_liter_gasoline: `${data.kapal.minimum_knot_per_liter_gasoline} KTS`,
    maximum_knot_per_liter_gasoline: `${data.kapal.maximum_knot_per_liter_gasoline} KTS`,
    history_per_second: data.kapal.history_per_second,
    record_status: data.kapal.record_status ? "Active" : "Inactive",
    created_at: new Date(data.kapal.created_at).toLocaleString(),
    updated_at: new Date(data.kapal.updated_at).toLocaleString(),
  };

  for (const [id, value] of Object.entries(elements)) {
    const element = document.getElementById(id);
    if (element) {
      if (id === "status_telnet") {
        element.textContent = value;
        element.style.color = value === "Connected" ? "green" : "red";
      } else {
        element.textContent = value;
      }
    }
  }

  // Update vessel image
  const imageElement = document.getElementById("vessel-image");
  if (imageElement && data.kapal && data.kapal.image) {
    imageElement.src = "/public/upload/assets/image/vessel/" + data.kapal.image;
    imageElement.alt = `Image of ${device}`;
  }

  //   updateVesselImage(
  //     data.kapal.image_map,
  //     data.kapal.width_m,
  //     data.kapal.height_m,
  //     data.kapal.top_range,
  //     data.kapal.left_range
  // );
}

// TODO : Image Detail Vessel That displayed on MAP locate the GPS with Dot with real places
// function updateVesselImage(imageUrl, width, height, topRange, leftRange) {
//   const container = document.getElementById('vessel-image-container');
//   container.innerHTML = ''; // Clear previous content

//   // Set container height based on aspect ratio (width is already 100%)
//   container.style.paddingBottom = `${(width / height) * 100}%`;

//   const imageWrapper = document.createElement('div');
//   imageWrapper.style.position = 'absolute';
//   imageWrapper.style.top = '0';
//   imageWrapper.style.left = '0';
//   imageWrapper.style.width = '100%';
//   imageWrapper.style.height = '100%';
//   imageWrapper.style.transform = 'rotate(90deg)';
//   imageWrapper.style.transformOrigin = 'top left';

//   const img = document.createElement('img');
//   img.src = `/public/upload/assets/image/vessel_map/${imageUrl}`;
//   img.style.width = '100%';
//   img.style.height = '100%';
//   img.style.objectFit = 'contain';
//   img.style.transform = 'rotate(-90deg) translateY(-100%)';
//   img.style.transformOrigin = 'top left';

//   const dot = document.createElement('div');
//   dot.style.position = 'absolute';
//   dot.style.width = '6px';
//   dot.style.height = '6px';
//   dot.style.borderRadius = '50%';
//   dot.style.backgroundColor = 'red';
//   dot.style.border = '1px solid white';

//   // Calculate dot position (swapped due to rotation)
//   const topPercentage = (leftRange / width) * 100;
//   const leftPercentage = (1 - (topRange / height)) * 100;

//   dot.style.top = `${topPercentage}%`;
//   dot.style.left = `${leftPercentage}%`;
//   dot.style.transform = 'translate(-50%, -50%)';

//   imageWrapper.appendChild(img);
//   container.appendChild(imageWrapper);
//   container.appendChild(dot);
// }

function formatWaterDepthNumber(number) {
  const [part1, part2] = number
    .toString()
    .padStart(3, "0")
    .match(/^(\d+)(\d{2})$/)
    .slice(1);
  return parseFloat(`${part1 || "0"}.${part2}`);
}

function getDataKapalMarker(device) {
  const vessel_record_preview = document.getElementById(
    "vessel_record_preview"
  );
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
  progressSlider.value =
    progressSlider.max =
    totalVesselHistoryRecords =
    currentAnimationIndex =
      0;
  if (isAnimationPlaying) stopVesselHistoryAnimation();
}

// Initialize WebSocket connection
connectWebSocket();
