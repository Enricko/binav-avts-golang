// Constants
const ANIMATION_DURATION = 75000; // 75 seconds in milliseconds

// Global variables
let previewTimeoutID;
let isPreviewActive = false;
let vesselHistoryData = [];
let vesselHistoryDatetime = [];
let totalVesselHistoryRecords = 0;

let vesselPolylineHistory;
let isAnimationPlaying = false;
let animationTimeoutID;
let shouldStopAnimation = false;
let historyMarker;
let percentage;

let startDatetimeFilter;
let endDatetimeFilter;

let currentAnimationIndex = 0; // Track the current index of animation
let animationFrameDuration = ANIMATION_DURATION / totalVesselHistoryRecords; // Duration of animation per frame

// DOM Elements
const progressSlider = document.getElementById("progress-slider1");
const btnPlay = document.getElementById("play-animation");
const btnLoad = document.getElementById("load-vessel-history");
const loadingSpinner = document.getElementById("spinner");
const btnDownloadCSV = document.getElementById("history-download-csv");
const sidebar = document.getElementById("detail-vessel");
const submitFilter = document.getElementById("submitFilter");
const startDateTimeInput = document.getElementById("start-date-time");
const endDateTimeInput = document.getElementById("end-date-time");
const filterModal = document.getElementById("filterModal");
const modalInstance = new bootstrap.Modal(filterModal);

// Event Listeners
document.addEventListener("DOMContentLoaded", () => {
  if (progressSlider) {
    progressSlider.addEventListener("input", updateHistorybySlider);
  }
  if (btnPlay) btnPlay.addEventListener("click", togglePlayPause);
  if (btnLoad)
    btnLoad.addEventListener("click", () => {
      loadVesselHistoryData(startDatetimeFilter, endDatetimeFilter);
    });
  if (submitFilter)
    submitFilter.addEventListener("click", function (event) {
      startEndDatetimeFilterForm();
      modalInstance.hide();
    });
  if (btnDownloadCSV)
    btnDownloadCSV.addEventListener("click", () => {
      downloadCSV(
        `${currentSelectedMarker}_record_${formatDateTime(
          startDatetimeFilter
        )}_to_${formatDateTime(endDatetimeFilter)}.csv`,
        vesselHistoryData.map((data) => data.record)
      );
    });
});

$("#filterModal").on("hidden.bs.modal", function () {
  startEndDatetimeFilterForm();
});

// Functions
function createPreviewButton(map) {
  const previewButton = document.createElement("button");
  previewButton.id = "vessel_record_preview";
  previewButton.classList.add("btn", "btn-primary", "rounded-circle", "ml-4");
  previewButton.style.cssText = `
    background-color: white;
    border: 0;
    width: 50px;
    height: 50px;
    margin-right: 0.5rem;
    margin-bottom: 0.5rem;
    display: none;
  `;
  previewButton.innerHTML = '<i class="fas fa-eye" style="color: black;"></i>';
  previewButton.title = "Preview Vessel Data";
  previewButton.addEventListener("click", toggleVesselDetailSidebar);

  return previewButton;
}

function initializeCompleteHistory(payload) {
  loadingSpinner.style.display = "none";

  btnPlay.disabled = btnDownloadCSV.disabled = totalVesselHistoryRecords === 0;
  progressSlider.value = 0;
  progressSlider.max = totalVesselHistoryRecords;
  animationFrameDuration = ANIMATION_DURATION / totalVesselHistoryRecords;
  currentAnimationIndex = 0;

  console.log(totalVesselHistoryRecords);
  console.log(vesselHistoryData);
}

function toggleVesselDetailSidebar() {
  if (isPreviewActive) {
    sidebar.style.display = "block";
    previewTimeoutID = setInterval(() => {
      dataKapalMarker(currentSelectedMarker);
    }, 100);
  } else {
    sidebar.style.display = "none";
    clearInterval(previewTimeoutID);
  }
  isPreviewActive = !isPreviewActive;
}

function displayVesselHistoryPolyline() {
  if (vesselHistoryData.length === 0) return;

  vesselPolylineHistory = [];
  vesselPolylineHistory.forEach((polyline) => polyline.setMap(null));

  let currentSegment = [];
  let currentColor = getPolylineColor(vesselHistoryData[0].status);

  vesselHistoryData.forEach((data, index) => {
    currentSegment.push(data.latlng);

    createPolylineSegment(currentSegment, currentColor);
    currentSegment = [data.latlng];
    currentColor = getPolylineColor(data.status);
  });

  if (currentSegment.length > 0) {
    createPolylineSegment(currentSegment, currentColor);
  }

  map.setCenter(vesselHistoryData[vesselHistoryData.length - 1].latlng);
}

function createPolylineSegment(path, color) {
  const polyline = new google.maps.Polyline({
    path: path,
    geodesic: true,
    strokeColor: color,
    strokeOpacity: 1.0,
    strokeWeight: 2,
  });
  polyline.setMap(map);
  vesselPolylineHistory.push(polyline);
}

function getPolylineColor(telnetStatus) {
  return telnetStatus === "Connected" ? "#0077b6" : "#ff0000";
}

function playVesselHistoryAnimation() {
  if (shouldStopAnimation) {
    shouldStopAnimation = false;
  } else if (currentAnimationIndex === 0) {
    initializeHistoryMarker();
  }
  animateMarker();
}

function initializeHistoryMarker() {
  if (!vesselHistoryData.length) return;

  if (!historyMarker) {
    historyMarker = new VesselOverlayHistory(
      map,
      vesselHistoryData[currentAnimationIndex].latlng,
      vesselHistoryData.kapal.top_range,
      vesselHistoryData.kapal.left_range,
      vesselHistoryData.kapal.width_m,
      vesselHistoryData.kapal.height_m,
      (vesselHistoryData[currentAnimationIndex].record.heading_degree +
        vesselHistoryData.kapal.calibration +
        vesselHistoryData.kapal.heading_direction) %
        360,
      vesselHistoryData.kapal.image_map
    );
  }
}

function animateMarker() {
  if (shouldStopAnimation) {
    isAnimationPlaying = false;
    clearTimeout(animationTimeoutID);
    return;
  }

  if (currentAnimationIndex >= totalVesselHistoryRecords) {
    stopVesselHistoryAnimation();
    return;
  }

  const position = vesselHistoryData[currentAnimationIndex].latlng;

  map.setCenter(position);
  historyMarker.update(
    position,
    vesselHistoryData.kapal.top_range,
    vesselHistoryData.kapal.left_range,
    vesselHistoryData.kapal.width_m,
    vesselHistoryData.kapal.height_m,
    (vesselHistoryData[currentAnimationIndex].record.heading_degree +
      vesselHistoryData.kapal.calibration +
      vesselHistoryData.kapal.heading_direction) %
      360,
    vesselHistoryData.kapal.image_map
  );
  progressSlider.value = currentAnimationIndex;
  updateHistoryTable(currentAnimationIndex);
  currentAnimationIndex++;
  animationTimeoutID = setTimeout(animateMarker, animationFrameDuration);
}

function stopVesselHistoryAnimation() {
  shouldStopAnimation = true;
  clearTimeout(animationTimeoutID);
  isAnimationPlaying = false;
  updatePlayPauseButton();
}

function togglePlayPause() {
  isAnimationPlaying = !isAnimationPlaying;
  if (isAnimationPlaying) {
    playVesselHistoryAnimation();
  } else {
    stopVesselHistoryAnimation();
  }
  updatePlayPauseButton();
}

function updatePlayPauseButton() {
  btnPlay.innerHTML = isAnimationPlaying
    ? '<i class="fas fa-pause" style="color: rgb(255, 255, 255);"></i> Pause'
    : '<i class="fas fa-play" style="color: rgb(255, 255, 255);"></i> Play';
}

function loadVesselHistoryData(datetimeFrom, datetimeTo) {
  const message = {
      type: "vessel_records_request",
      payload: {
          call_sign: currentSelectedMarker,
          start: datetimeFrom,
          end: datetimeTo
      }
  };
  websocket.send(JSON.stringify(message));
}

function updateHistorybySlider() {
  let progressValue = parseInt(progressSlider.value, 10);
  if (progressValue >= 0 && progressValue < totalVesselHistoryRecords) {
    if (!historyMarker) {
      initializeHistoryMarker();
    } else {
      const position = vesselHistoryData[progressValue].latlng;

      map.setCenter(position);
      historyMarker.update(
        position,
        vesselHistoryData.kapal.top_range,
        vesselHistoryData.kapal.left_range,
        vesselHistoryData.kapal.width_m,
        vesselHistoryData.kapal.height_m,
        (vesselHistoryData[progressValue].record.heading_degree +
          vesselHistoryData.kapal.calibration +
          vesselHistoryData.kapal.heading_direction) %
          360,
        vesselHistoryData.kapal.image_map
      );
    }
    updateHistoryTable(progressValue);
    currentAnimationIndex = progressValue;
  }
}

function updateHistoryTable(index) {
  if (index < 0 || index >= vesselHistoryData.length) return;

  const record = vesselHistoryData[index].record;
  document.getElementById("record_of_vessel").textContent = `${
    index + 1
  } of ${totalVesselHistoryRecords} Records`;
  document.getElementById("latitude_record").textContent = record.latitude;
  document.getElementById("longitude_record").textContent = record.longitude;
  document.getElementById("heading_hdt_record").textContent = `${
    record.heading_degree + vesselHistoryData.kapal.calibration
  }°`;
  document.getElementById(
    "SOG_record"
  ).textContent = `${record.speed_in_knots} KTS`;
  document.getElementById("SOLN_record").textContent =
    record.gps_quality_indicator;
  document.getElementById("datetime_record").textContent =
    formatDateTimeDisplay(record.created_at);
  document.getElementById(
    "water_depth_record"
  ).textContent = `${formatWaterDepthNumber(record.water_depth)} Meter`;
}

function defaultHistoryTable() {
  const elements = {
    record_of_vessel: "0 of 0 Records",
    latitude_record: "-",
    longitude_record: "-",
    heading_hdt_record: "-°",
    SOG_record: "- KTS",
    SOLN_record: "-",
    datetime_record: "-",
    water_depth_record: "- Meter",
  };

  for (const [id, value] of Object.entries(elements)) {
    if (document.getElementById(id))
      document.getElementById(id).textContent = value;
  }
}

function filterByDateRange(data, startDate, endDate) {
  const start = new Date(startDate);
  const end = new Date(endDate);
  return data.filter((item) => {
    const createdAt = new Date(item.created_at);
    return createdAt >= start && createdAt <= end;
  });
}

function formatToISO(dateString) {
  const date = new Date(dateString);
  const offset = date.getTimezoneOffset();
  const offsetHours = String(Math.floor(Math.abs(offset) / 60)).padStart(
    2,
    "0"
  );
  const offsetMinutes = String(Math.abs(offset) % 60).padStart(2, "0");
  const sign = offset > 0 ? "-" : "+";

  return (
    date.toISOString().slice(0, 19) + sign + offsetHours + ":" + offsetMinutes
  );
}

function startEndDatetimeFilterForm() {
  const start = new Date(startDateTimeInput.value);
  const end = new Date(endDateTimeInput.value);

  startDatetimeFilter = formatToISO(start);
  endDatetimeFilter = formatToISO(end);

  document.getElementById("filter_history_start").textContent =
    formatDateWithMidnight(start);
  document.getElementById("filter_history_end").textContent = formatDate(end);

  if (endDatetimeFilter <= startDatetimeFilter) {
    alert("End date and time must be after start date and time.");
  }
}

function downloadCSV(filename, data) {
  btnDownloadCSV.disabled = true;
  loadingSpinner.style.display = "block";

  const headers = [
    "time",
    "latitude",
    "longitude",
    "heading_degree",
    "speed_in_knots",
    "gps_quality_indicator",
    "water_depth",
  ];
  const headerMapping = {
    time: "created_at",
    latitude: "latitude",
    longitude: "longitude",
    heading_degree: "heading_degree",
    speed_in_knots: "speed_in_knots",
    gps_quality_indicator: "gps_quality_indicator",
    water_depth: "water_depth",
  };

  const csvRows = [headers.join(",")];
  let index = 0;

  function processNextChunk() {
    const chunkSize = 1000;
    for (let i = 0; i < chunkSize && index < data.length; i++, index++) {
      const row = data[index];
      const values = headers.map((header) => {
        let value = row[headerMapping[header]];

        if (header === "water_depth" && typeof value === "number") {
          value = formatWaterDepthNumber(value);
        }

        if (header === "latitude" || header === "longitude") {
          value = value.replace(/Â°/g, "°").replace(/"/g, '""');
        }

        if (header === "heading_degree") {
          value = value + vesselHistoryData.kapal.calibration;
        }

        if (header === "time") {
          value = formatDateTimeDisplay(value);
        }

        return value !== undefined ? value : "";
      });

      csvRows.push(values.join(","));
    }

    if (index < data.length) {
      setTimeout(processNextChunk, 0);
    } else {
      const csvString = csvRows.join("\n");
      const bom = "\uFEFF"; // UTF-8 BOM
      const blob = new Blob([bom + csvString], {
        type: "text/csv;charset=utf-8;",
      });
      const url = URL.createObjectURL(blob);

      const link = document.createElement("a");
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      btnDownloadCSV.disabled = false;
      loadingSpinner.style.display = "none";
    }
  }

  processNextChunk();
}
