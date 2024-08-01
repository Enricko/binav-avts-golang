let timeoutIDDetailData;
let isPreview = false;
let historyVesselData = [];
let totalHistoryVesselData = 0;

let vesselPolylineHistory;
let markerStrava;

let isPlayingAnimation = false;

let timeoutIdAnimation;
let stopAnimation = false;

const durationAnimation = 60000; // 60 seconds in milliseconds

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
      display:none;
    `;
  controlButton.innerHTML =
    '<i class="fas fa-solid fa-eye" style="color: black;"></i>';
  controlButton.title = "Preview Button";
  controlButton.addEventListener("click", viewDetailKapal);

  return controlButton;
}

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

function viewHistoryPolyline() {
  vesselPolylineHistory = new google.maps.Polyline({
    path: historyVesselData,
    geodesic: true,
    strokeColor: "#FF0000",
    strokeOpacity: 1.0,
    strokeWeight: 2,
  });
  vesselPolylineHistory.setMap(map);
}

function playStrava() {
  if (!isPlayingAnimation) {
    isPlayingAnimation = true;
    markerStrava = new google.maps.Marker({
      position: historyVesselData[0],
      map: map,
      title: "Marker",
    });
    let index = 0;
    const interval = durationAnimation / totalHistoryVesselData;
    function animateMarker() {
      if (stopAnimation) {
        isPlayingAnimation = false;
        stopAnimation = false;
        return;
      }
      markerStrava.setPosition(historyVesselData[index]);
      map.panTo(historyVesselData[index]); // Center the map on the marker's position
      index++;
      if (index < totalHistoryVesselData) {
        timeoutIdAnimation = setTimeout(animateMarker, interval);
      }
      if (index >= totalHistoryVesselData) {
        isPlayingAnimation = false;
      }
    }
    animateMarker();
  }
}

function dismissStrava() {
  stopAnimation = true;
  clearTimeout(timeoutIdAnimation);
  isPlayingAnimation = false;
}

async function loadVesselHistory() {
  const loading = document.getElementById("spinner");
  loading.style.display = "block";

  const url = `http://127.0.0.1:8080/vessel_records/${currentSelectedMarker}`;
  try {
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error("Network response was not ok " + response.statusText);
    }

    const result = await response.json();

    historyVesselData = result.records.map((record) => {
      return {
        lat: convertDMSToDecimal(record.latitude),
        lng: convertDMSToDecimal(record.longitude),
      };
    });
    totalHistoryVesselData = result.total_record;
    document.getElementById("total_records").innerHTML = totalHistoryVesselData;
    loading.style.display = "none";
    if (totalHistoryVesselData > 0) {
      btnPlay.disabled = false;
    } else {
      btnPlay.disabled = true;
    }
    viewHistoryPolyline();
  } catch (error) {
    console.error("There was a problem with the fetch operation:", error);
  }
}

const btnPlay = document.getElementById("play-animation");
btnPlay.addEventListener("click", playStrava);

const btnLoad = document.getElementById("load-vessel-history");
btnLoad.addEventListener("click", loadVesselHistory);
