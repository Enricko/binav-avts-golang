// Variables
let timeoutIDDetailData;
let isPreview = false;
let historyVesselData = [];
let totalHistoryVesselData = 0;

let vesselPolylineHistory;
let markerStrava;
let isPlayingAnimation = false;
let timeoutIdAnimation;
let stopAnimation = false;
let markerHistory;

const durationAnimation = 75000; // 60 seconds in milliseconds

let animationIndex = 0;  // Track the current index of animation
let animationInterval = durationAnimation / totalHistoryVesselData; // Duration of animation per frame

// Functions

// Create preview button
function createPreviewButton(map) {
  const controlButton = document.createElement("button");
  controlButton.id = "vessel_record_preview";
  controlButton.classList.add("btn", "btn-primary", "rounded-circle", "ml-4");
  controlButton.style.cssText = `
      background-color: white;
      border: 0;
      width: 50px;
      height: 50px;
      margin-right: 0.5rem;
      margin-bottom: 0.5rem;
      display: none;
    `;
  controlButton.innerHTML = '<i class="fas fa-solid fa-eye" style="color: black;"></i>';
  controlButton.title = "Preview Button";
  controlButton.addEventListener("click", viewDetailKapal);

  return controlButton;
}

// Toggle vessel detail sidebar
function viewDetailKapal() {
  const sidebarDetail = document.getElementById("detail-vessel");
  if (isPreview) {
    sidebarDetail.style.display = "block";
    timeoutIDDetailData = setInterval(() => {
      dataKapalMarker(currentSelectedMarker);
    }, 100);
  } else {
    sidebarDetail.style.display = "none";
    clearInterval(timeoutIDDetailData);
  }
  isPreview = !isPreview;
}

// Preview Strava marker animation
function previewStrava() {
  let index = 0;
  function animateMarker() {
    markerStrava.setPosition(historyVesselData[index]);
    index++;
    if (index < historyVesselData.length) {
      setTimeout(animateMarker, 500);
    }
  }
  animateMarker();
}

// View history polyline
function viewHistoryPolyline() {
  const latlngArray = historyVesselData.map((data) => data.latlng);
  vesselPolylineHistory = new google.maps.Polyline({
    path: latlngArray,
    geodesic: true,
    strokeColor: "#FF0000",
    strokeOpacity: 1.0,
    strokeWeight: 2,
  });
  vesselPolylineHistory.setMap(map);
}

// Play Strava animation
function playHistory() {
  if (!isPlayingAnimation) {
    animationIndex = 0;
    isPlayingAnimation = true;
    if (stopAnimation) {
      stopAnimation = false;
    } else {
      // Initialize markerHistory if not already created
      markerHistory = new VesselOverlayHistory(
        map,
        {
          lat: historyVesselData[animationIndex].latlng.lat,
          lng: historyVesselData[animationIndex].latlng.lng,
        },
        historyVesselData.kapal.top_range,
        historyVesselData.kapal.left_range,
        historyVesselData.kapal.width_m,
        historyVesselData.kapal.height_m,
        (historyVesselData[animationIndex].record.heading_degree +
          historyVesselData.kapal.calibration +
          historyVesselData.kapal.heading_direction) % 360,
        historyVesselData.kapal.image_map
      );
    }
    animateMarker();
  }
}

function animateMarker() {
  if (stopAnimation) {
    isPlayingAnimation = false;
    clearTimeout(timeoutIdAnimation);
    return;
  }
  const position = {lat : historyVesselData[animationIndex].latlng.lat, lng : historyVesselData[animationIndex].latlng.lng};

  map.setCenter(position);
  markerHistory.update(
    position,
    historyVesselData.kapal.top_range,
    historyVesselData.kapal.left_range,
    historyVesselData.kapal.width_m,
    historyVesselData.kapal.height_m,
    (historyVesselData[animationIndex].record.heading_degree +
      historyVesselData.kapal.calibration +
      historyVesselData.kapal.heading_direction) % 360,
    historyVesselData.kapal.image_map
  );
  displayTableHistory(animationIndex);
  animationIndex++;
  if (animationIndex < totalHistoryVesselData) {
    timeoutIdAnimation = setTimeout(animateMarker, animationInterval);
  } else {
    isPlayingAnimation = false;
  }
}

// Stop Strava animation
function dismissHistory() {
  stopAnimation = true;
  clearTimeout(timeoutIdAnimation);
  isPlayingAnimation = false;
}

// Load vessel history data
async function loadVesselHistory() {
  console.log(historyVesselData);
  const loading = document.getElementById("spinner");
  loading.style.display = "block";

  const url = `http://127.0.0.1:8080/vessel_records/${currentSelectedMarker}`;
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error("Network response was not ok " + response.statusText);
    }

    const result = await response.json();
    historyVesselData = result.records.map((record) => ({
      record: record,
      latlng: {
        lat: convertDMSToDecimal(record.latitude),
        lng: convertDMSToDecimal(record.longitude),
      },
    }));
    historyVesselData["kapal"] = result.kapal;
    totalHistoryVesselData = result.total_record;
    document.getElementById("total_records").innerHTML = totalHistoryVesselData;
    loading.style.display = "none";
    btnPlay.disabled = totalHistoryVesselData === 0;
    viewHistoryPolyline();
  } catch (error) {
    console.error("There was a problem with the fetch operation:", error);
  }
}

// Display vessel history table data
function displayTableHistory(index) {
  document.getElementById("record_of_vessel").textContent = `${index + 1} of ${totalHistoryVesselData} Records`;
  const record = historyVesselData[index].record;
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
btnPlay.addEventListener("click", playHistory);

const btnLoad = document.getElementById("load-vessel-history");
btnLoad.addEventListener("click", loadVesselHistory);
