// Variables
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

const animationDuration = 75000; // 75 seconds in milliseconds
let currentAnimationIndex = 0; // Track the current index of animation
let animationFrameDuration = animationDuration / totalVesselHistoryRecords; // Duration of animation per frame

// DOM Elements
const progressSlider = document.getElementById("progress-slider1");
const btnPlay = document.getElementById("play-animation");
const btnLoad = document.getElementById("load-vessel-history");
const loadingSpinner = document.getElementById("spinner");

// Ensure DOM is fully loaded before attaching event listeners
document.addEventListener("DOMContentLoaded", () => {
  progressSlider.addEventListener("input", updateHistorybySlider);
  btnPlay.addEventListener("click", togglePlayPause);
  btnLoad.addEventListener("click", () => {
    loadVesselHistoryData(
      startDatetimeFilter,
      endDatetimeFilter
    );
  });
});

// Create preview button
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

// Toggle vessel detail sidebar
function toggleVesselDetailSidebar() {
  const sidebar = document.getElementById("detail-vessel");
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

// View history polyline
function displayVesselHistoryPolyline() {
  if (vesselHistoryData.length === 0) return;

  const latlngArray = vesselHistoryData.map((data) => data.latlng);
  console.log("latlngArray");
  console.log(latlngArray);
  
  
  vesselPolylineHistory = new google.maps.Polyline({
    path: latlngArray,
    geodesic: true,
    strokeColor: "#FF0000",
    strokeOpacity: 1.0,
    strokeWeight: 2,
  });
  vesselPolylineHistory.setMap(map);
}

// Play vessel history animation
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

  historyMarker = new VesselOverlayHistory(
    map,
    {
      lat: vesselHistoryData[currentAnimationIndex].latlng.lat,
      lng: vesselHistoryData[currentAnimationIndex].latlng.lng,
    },
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

  const position = {
    lat: vesselHistoryData[currentAnimationIndex].latlng.lat,
    lng: vesselHistoryData[currentAnimationIndex].latlng.lng,
  };

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

// Stop vessel history animation
function stopVesselHistoryAnimation() {
  shouldStopAnimation = true;
  clearTimeout(animationTimeoutID);
  isAnimationPlaying = false;
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
  if (isAnimationPlaying) {
    btnPlay.innerHTML = `<i class="fas fa-pause" style="color: rgb(255, 255, 255);"></i>`;
  } else {
    btnPlay.innerHTML = `<i class="fas fa-play" style="color: rgb(255, 255, 255);"></i>`;
  }
}

// Load vessel history data
async function loadVesselHistoryData(datetimeFrom, datetimeTo) {
  loadingSpinner.style.display = "block";

  console.log("datetime");
  console.log(datetimeFrom);
  console.log(datetimeTo);
  
  const url = `http://127.0.0.1:8080/vessel_records/${currentSelectedMarker}`;

  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error("Network response was not ok " + response.statusText);
    }

    const result = await response.json();

    if (datetimeFrom != null && datetimeTo != null) {
      console.log("with filter");
      const filteredData = filterByDateRange(
        result.records,
        datetimeFrom,
        datetimeTo
      );
      
      vesselHistoryData = filteredData.map((record) => ({
        record: record,
        latlng: {
          lat: convertDMSToDecimal(record.latitude),
          lng: convertDMSToDecimal(record.longitude),
        },
        dateTime: record.created_at,
      }));
      console.log("hhhhhh");
      
      console.log(vesselHistoryData);
    }else{
      console.log("without filter");
      
      vesselHistoryData = result.records.map((record) => ({
        record: record,
        latlng: {
          lat: convertDMSToDecimal(record.latitude),
          lng: convertDMSToDecimal(record.longitude),
        },
        dateTime: record.created_at,
      }));
    }

    // const sortedData = await sortByCreatedAt(vesselHistoryData);
    // console.log(sortedData);
   
    vesselHistoryData["kapal"] = result.kapal;
    totalVesselHistoryRecords = vesselHistoryData.length;
    document.getElementById("total_records").textContent =
      totalVesselHistoryRecords;
    loadingSpinner.style.display = "none";
    btnPlay.disabled = totalVesselHistoryRecords === 0;
    progressSlider.value = 0;
    progressSlider.max = totalVesselHistoryRecords;
    animationFrameDuration = animationDuration / totalVesselHistoryRecords; // Update frame duration
    displayVesselHistoryPolyline();
  } catch (error) {
    console.error("There was a problem with the fetch operation:", error);
    loadingSpinner.style.display = "none";
  }
}

function updateHistorybySlider() {
  let progressValue = parseInt(progressSlider.value, 10);
  if (progressValue >= 0 && progressValue < totalVesselHistoryRecords) {
    if (!historyMarker) {
      initializeHistoryMarker();
    } else {
      const position = {
        lat: vesselHistoryData[progressValue].latlng.lat,
        lng: vesselHistoryData[progressValue].latlng.lng,
      };

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

// Update vessel history table
function updateHistoryTable(index) {
  if (index < 0 || index >= vesselHistoryData.length) return;

    document.getElementById("record_of_vessel").textContent = `${index + 1} of ${totalVesselHistoryRecords} Records`;
    const record = vesselHistoryData[index].record;
    // percentage = ((index + 1) / totalVesselHistoryRecords) * 100;
    // progressSlider.value = percentage;
    document.getElementById("latitude_record").textContent = record.latitude;
    document.getElementById("longitude_record").textContent = record.longitude;
    document.getElementById("heading_hdt_record").textContent = record.heading_degree;
    document.getElementById("SOG_record").textContent = `${record.speed_in_knots} KTS`;
    document.getElementById("SOLN_record").textContent = record.gps_quality_indicator;
    document.getElementById("datetime_record").textContent = formatDateTime(record.created_at);
    document.getElementById("water_depth_record").textContent = `${formatWaterDepthNumber(record.water_depth)} Meter`;
}


// submitDateTime.addEventListener('click', function (event) {
//   formDateTime.submit();
// });

// function sortByCreatedAt(data) {
//   return data.sort((a, b) => new Date(a.dateTime) - new Date(b.dateTime));
// }

function filterByDateRange(data, startDate, endDate) {
  const start = new Date(startDate);
  const end = new Date(endDate);
  return data.filter((item) => {
    const createdAt = new Date(item.created_at);
    return createdAt >= start && createdAt <= end;
  });
}

// function extractLatLon(filteredData) {
//   return filteredData.map((item) => {
//     return {
//       latitude: convertDMSToDecimal(item.latitude),
//       longitude: convertDMSToDecimal(item.longitude),
//     };
//   });
// }
function formatToISO(dateString) {
  // Parse the date string to a Date object
  const date = new Date(dateString);

  // Extract components
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  const offset = date.getTimezoneOffset();
  
  // Calculate the timezone offset in hours and minutes
  const offsetHours = String(Math.floor(Math.abs(offset) / 60)).padStart(2, '0');
  const offsetMinutes = String(Math.abs(offset) % 60).padStart(2, '0');
  const sign = offset > 0 ? '-' : '+';
  
  // Construct ISO 8601 string
  const isoString = `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${sign}${offsetHours}:${offsetMinutes}`;
  
  return isoString;
}

const submitFilter = document.getElementById("submitFilter");
const startDateTimeInput = document.getElementById("start-date-time");
const endDateTimeInput = document.getElementById("end-date-time");
const filterModal = document.getElementById("filterModal");
const modalInstance = new bootstrap.Modal(filterModal);

submitFilter.addEventListener("click", function (event) {
  const start = new Date(startDateTimeInput.value);
  const end = new Date(endDateTimeInput.value);

  startDatetimeFilter = formatToISO(start);
  endDatetimeFilter = formatToISO(end);

  console.log("Success");
  console.log(startDatetimeFilter);
  console.log(endDatetimeFilter);

  if (endDatetimeFilter <= startDatetimeFilter) {
    event.preventDefault();
    alert("End date and time must be after start date and time.");
  }
  modalInstance.hide();
  
});
