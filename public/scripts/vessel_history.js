// Variables
let previewTimeoutID;
let isPreviewActive = false;
let vesselHistoryData = [];
let totalVesselHistoryRecords = 0;

let vesselPolylineHistory;
let isAnimationPlaying = false;
let animationTimeoutID;
let shouldStopAnimation = false;
let historyMarker;
let percentage;

const animationDuration = 75000; // 75 seconds in milliseconds

let currentAnimationIndex = 0; // Track the current index of animation
let animationFrameDuration = animationDuration / totalVesselHistoryRecords; // Duration of animation per frame

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
  const latlngArray = vesselHistoryData.map(data => data.latlng);
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
      vesselHistoryData.kapal.heading_direction) % 360,
    vesselHistoryData.kapal.image_map
  );
}

function animateMarker() {
  if (shouldStopAnimation) {
    isAnimationPlaying = false;
    clearTimeout(animationTimeoutID);
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
      vesselHistoryData.kapal.heading_direction) % 360,
    vesselHistoryData.kapal.image_map
  );
  updateHistoryTable(currentAnimationIndex);
  currentAnimationIndex++;
  if (currentAnimationIndex < totalVesselHistoryRecords) {
    animationTimeoutID = setTimeout(animateMarker, animationFrameDuration);
  } else {
    stopVesselHistoryAnimation();
    currentAnimationIndex = 0;
  }
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
async function loadVesselHistoryData() {
  console.log(vesselHistoryData);
  const loadingSpinner = document.getElementById("spinner");
  loadingSpinner.style.display = "block";

  const url = `http://127.0.0.1:8080/vessel_records/${currentSelectedMarker}`;
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error("Network response was not ok " + response.statusText);
    }

    const result = await response.json();
    vesselHistoryData = result.records.map(record => ({
      record: record,
      latlng: {
        lat: convertDMSToDecimal(record.latitude),
        lng: convertDMSToDecimal(record.longitude),
      },
    }));
    vesselHistoryData["kapal"] = result.kapal;
    totalVesselHistoryRecords = result.total_record;
    document.getElementById("total_records").textContent = totalVesselHistoryRecords;
    loadingSpinner.style.display = "none";
    btnPlay.disabled = totalVesselHistoryRecords === 0;
    displayVesselHistoryPolyline();
  } catch (error) {
    console.error("There was a problem with the fetch operation:", error);
  }
}
const progressSlider = document.getElementById('progress-slider1');

progressSlider.addEventListener("input", toggleVesselDetailSidebar);

function updateHistorybySlider(){

}



// progressSlider.addEventListener('input', (event) => {
//   percentage = event.target.value;
// });

// Update vessel history table
function updateHistoryTable(index) {
  document.getElementById("record_of_vessel").textContent = `${index + 1} of ${totalVesselHistoryRecords} Records`;
  const record = vesselHistoryData[index].record;
  percentage = ((index + 1) / totalVesselHistoryRecords) * 100;
  progressSlider.value = percentage;
  document.getElementById("latitude_record").textContent = record.latitude;
  document.getElementById("longitude_record").textContent = record.longitude;
  document.getElementById("heading_hdt_record").textContent = record.heading_degree;
  document.getElementById("SOG_record").textContent = `${record.speed_in_knots} KTS`;
  document.getElementById("SOLN_record").textContent = record.gps_quality_indicator;
  document.getElementById("water_depth_record").textContent = record.water_depth;
  document.getElementById("datetime_record").textContent = formatDateTime(record.created_at);
}

// Event listeners
const btnPlay = document.getElementById("play-animation");
btnPlay.addEventListener("click", togglePlayPause);

const btnLoad = document.getElementById("load-vessel-history");
btnLoad.addEventListener("click", loadVesselHistoryData);

const video = document.getElementById('myVideo');


video.addEventListener('timeupdate', () => {
  

});