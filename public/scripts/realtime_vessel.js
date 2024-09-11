let currentDevices = [];
let markers = {};
let dataDevices = {};
let isCameraLocked = false;
let lockedVesselId = null;

function toggleCameraLock(isLocked) {
  isCameraLocked = isLocked;
  if (!isLocked) {
    lockedVesselId = null;
  } else {
    lockedVesselId = currentSelectedMarker;
  }
  updateLockCameraButton();
}

function updateLockCameraButton() {
  const lockCameraBtn = document.getElementById("lock-camera-btn");
  if (lockCameraBtn) {
    if (isCameraLocked) {
      lockCameraBtn.innerHTML = '<i class="fas fa-unlock"></i> Unlock Camera';
      lockCameraBtn.classList.replace("btn-primary", "btn-secondary");
    } else {
      lockCameraBtn.innerHTML = '<i class="fas fa-lock"></i> Lock Camera';
      lockCameraBtn.classList.replace("btn-secondary", "btn-primary");
    }
  }
}

// Marker functions
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
    if (isCameraLocked && device === lockedVesselId) {
      map.panTo(position);
    }
  }
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

// Vessel data functions
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

  const imageElement = document.getElementById("vessel-image");
  if (imageElement && data.kapal && data.kapal.image) {
    imageElement.src = "/public/upload/assets/image/vessel/" + data.kapal.image;
    imageElement.alt = `Image of ${device}`;
  }
}

function getDataKapalMarker(device) {
  if (currentSelectedMarker !== device) {
    currentSelectedMarker = device;
    resetVesselState();

    // If this is the first vessel selection or if the camera is already locked,
    // lock the camera to the newly selected vessel
    if (!isCameraLocked || lockedVesselId === null) {
      toggleCameraLock(true);
    } else {
      // If a different vessel is selected while the camera is locked,
      // switch the lock to the new vessel
      lockedVesselId = device;
    }

    const vessel_record_preview = document.getElementById(
      "vessel_record_preview"
    );
    vessel_record_preview.style.display = "block";
    isPreviewActive = true;
    toggleVesselDetailSidebar();
    defaultHistoryTable();
    if (document.getElementById("total_records"))
      document.getElementById("total_records").textContent = "0";

    dataKapalMarker(device);
    startToEndDatetimeFilterForm();
  }
}

async function resetVesselState() {
  if (btnDownloadCSV) {
    btnPlay.disabled = btnDownloadCSV.disabled = true;
  }

  await clearPolylines();
  if (historyMarker) {
    historyMarker.setMap(null);
    historyMarker = null;
  }
  resetVesselHistoryAnimation();
}

function resetVesselHistoryAnimation() {
  if (progressSlider)
    progressSlider.value =
      progressSlider.max =
      totalVesselHistoryRecords =
      currentAnimationIndex =
        0;
  if (isAnimationPlaying) stopVesselHistoryAnimation();
}

// Utility functions
function convertDMSToDecimal(degreeMinute) {
  const [degrees, minutes, direction] = degreeMinute
    .match(/(\d+)°(\d+\.\d+)°([NS|EW])/)
    .slice(1);
  let decimalDegrees = parseFloat(degrees) + parseFloat(minutes) / 60;
  return direction === "S" || direction === "W"
    ? -decimalDegrees
    : decimalDegrees;
}

function formatWaterDepthNumber(number) {
  const [part1, part2] = number
    .toString()
    .padStart(3, "0")
    .match(/^(\d+)(\d{2})$/)
    .slice(1);
  return parseFloat(`${part1 || "0"}.${part2}`);
}
