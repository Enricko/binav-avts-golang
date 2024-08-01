let map;
let poly;
let kmzLayers = [];
const initialZoom = 16;
let isRulerOn = false;
let distanceMarkers = [];
let clickListener;
let mouseMoveListener;
let cursorMarker;
let totalLength = 0;

function createRulerButton(map) {
  const controlButton = document.createElement("button");

  controlButton.classList.add("btn", "btn-primary", "rounded-circle", "ml-4");
  controlButton.style.cssText = `
    background-color: white;
    border: 0;
    width: 50px;
    height: 50px;
    margin-right: 0.5rem;
    
  `;
  controlButton.innerHTML =
    '<i class="fas fa-solid fa-ruler" style="color: black;"></i>';
  controlButton.title = "Ruler Button";
  controlButton.addEventListener("click", toggleRuler);

  return controlButton;
}

function createProfileButton(map) {
  const profileButton = document.createElement("button");
  profileButton.classList.add("btn", "btn-primary", "rounded-circle", "mb-2");
  profileButton.style.cssText = `
    background-color: white;
    border: 0;
    width: 50px;
    height: 50px;
    margin-right: 0.5rem;
    margin-top: 0.5rem;
  `;
  profileButton.innerHTML = '<i class="fas fa-user" style="color: black;"></i>';
  profileButton.setAttribute("data-bs-toggle", "modal");
  profileButton.setAttribute("data-bs-target", "#profilePage");
  profileButton.title = "Profile Button";
  return profileButton;
}
function toggleRuler() {
  isRulerOn = !isRulerOn;

  const measurementWindow = document.getElementById("measurement-window");
  measurementWindow.style.display = isRulerOn ? "block" : "none";

  if (isRulerOn) {
    poly = new google.maps.Polyline({
      strokeColor: "#000000",
      strokeOpacity: 1.0,
      strokeWeight: 3,
      map: map,
      clickable: false,
      zIndex: 999999, // Set a high zIndex for the polyline
    });
    clickListener = map.addListener("click", addRulerPoint);
    mouseMoveListener = map.addListener("mousemove", updateCursorMarker);
    cursorMarker = new google.maps.Marker({
      map: map,
      icon: { path: google.maps.SymbolPath.CIRCLE, scale: 5 },
      zIndex: 999998, // Just below the polyline
      clickable: false,
    });

    // Display initial marker with label "0.0m"
    const initialPoint = poly.getPath().getAt(0); // Get the first point
    displayDistanceLabel(initialPoint, 0); // Display the label "0.0m"
  } else {
    if (poly) poly.setMap(null);
    if (clickListener) google.maps.event.removeListener(clickListener);
    if (mouseMoveListener) google.maps.event.removeListener(mouseMoveListener);
    if (cursorMarker) cursorMarker.setMap(null);
    distanceMarkers.forEach((marker) => marker.setMap(null));
    distanceMarkers = [];
    totalLength = 0;
    updateMeasurementWindow();
  }
}

function initMap() {
  map = new google.maps.Map(document.getElementById("map"), {
    zoom: initialZoom,
    center: { lat: -1.0574568371666666, lng: 117.35121627983332 },
    disableDefaultUI: true,
    zoomControl: true,
    scaleControl: true,
  });

  const profileButton = createProfileButton(map);
  map.controls[google.maps.ControlPosition.RIGHT_TOP].push(profileButton);

  const rulerButton = createRulerButton(map);
  map.controls[google.maps.ControlPosition.RIGHT_BOTTOM].push(rulerButton);

  const previewButton = createPreviewButton(map);
  map.controls[google.maps.ControlPosition.RIGHT_BOTTOM].push(previewButton);

  fetchKMZFiles();
  connectWebSocket();
  onSearchVessel();
}

function updateCursorMarker(event) {
  if (!isRulerOn) return;

  cursorMarker.setPosition(event.latLng);

  const path = poly.getPath();
  if (path.getLength() > 0) {
    const lastPoint = path.getAt(path.getLength() - 1);
    const distance = google.maps.geometry.spherical.computeDistanceBetween(
      lastPoint,
      event.latLng
    );
    displayCursorDistanceLabel(event.latLng, distance);
  }
}

function displayCursorDistanceLabel(position, distance) {
  const labelText =
    distance >= 1000
      ? `${(distance / 1000).toFixed(2)} Km`
      : `${distance.toFixed(2)} M`;

  cursorMarker.setLabel({
    text: labelText,
    color: "white", // Text color
    fontSize: "12px",
    fontWeight: "bold",
    className: "distance-label",
  });
}

// Add custom styles for the label
const style = document.createElement("style");
style.innerHTML = `
  .distance-label-clicked {
    background-color: black;
    padding: 2px 4px;
    border-radius: 3px;
    border: 2px solid white;
    }
    .distance-label {
      background-color: black;
      padding: 2px 4px;
      border-radius: 3px;
      margin-top: -30px;
      border: 2px solid white;
  }
`;
document.head.appendChild(style);

function addRulerPoint(event) {
  if (!isRulerOn) return;

  const path = poly.getPath();
  path.push(event.latLng);

  if (path.getLength() > 1) {
    const lastPoint = path.getAt(path.getLength() - 2);
    const newPoint = path.getAt(path.getLength() - 1);
    const distance = google.maps.geometry.spherical.computeDistanceBetween(
      lastPoint,
      newPoint
    );
    console.log(path.getLength() + " points");
    console.log(path.getAt(path.getLength() - 1) + " points");
    totalLength += distance;
    displayDistanceLabel(newPoint, distance);
  } else {
    console.log(path.getLength() + " points");
    console.log(path.getAt(path.getLength()) + " points");
    displayDistanceLabel(path.getAt(path.getLength() - 1), 0);
  }

  updateMeasurementWindow();
}

function displayDistanceLabel(position, distance) {
  const labelText =
    distance >= 1000
      ? `${(distance / 1000).toFixed(2)} Km`
      : `${distance.toFixed(2)} M`;

  // Create a marker with a custom label
  const marker = new google.maps.Marker({
    position,
    map,
    icon: { path: google.maps.SymbolPath.CIRCLE, scale: 0 },
    label: {
      text: labelText,
      color: "white", // Text color
      fontSize: "12px",
      fontWeight: "bold",
      className: "distance-label-clicked",
    },
  });

  distanceMarkers.push(marker);
}

function updateMeasurementWindow() {
  const totalClicks = poly ? poly.getPath().getLength() : 0;
  const measurementInfo = document.getElementById("measurement-info");
  measurementInfo.textContent = `Total Length: ${(totalLength / 1000).toFixed(
    2
  )} Km`;
}

function smoothPanTo(latLng) {
  const panSteps = 30;
  const panDuration = 1000;
  const panInterval = panDuration / panSteps;

  const startLat = map.getCenter().lat();
  const startLng = map.getCenter().lng();
  const endLat = latLng.lat;
  const endLng = latLng.lng;

  let step = 0;

  function panStep() {
    const lat = startLat + (endLat - startLat) * (step / panSteps);
    const lng = startLng + (endLng - startLng) * (step / panSteps);
    map.setCenter({ lat, lng });

    if (step < panSteps) {
      step++;
      setTimeout(panStep, panInterval);
    }
  }

  panStep();
}

function fetchKMZFiles() {
  fetch("/mappings")
    .then((response) => response.json())
    .then((data) =>
      data.forEach((mapping) => mapping.status && loadKMZLayer(mapping.file))
    )
    .catch((error) => console.error("Error fetching mappings:", error));
}

function getBaseURL() {
  // Get the full URL
  const fullURL = window.location.href;
  // Get the protocol and hostname
  const protocol = window.location.protocol;
  const host = window.location.host;
  // Combine them to get the base URL
  const baseURL = `${protocol}//${host}/`;
  return baseURL;
}

function loadKMZLayer(filePath) {
  // const kmzUrl = `${getBaseURK()}public/${filePath}`;
  const kmzUrl = `http://8.222.190.213/public/${filePath}`;
  const kmzLayer = new google.maps.KmlLayer({
    url: kmzUrl,
    map,
    preserveViewport: true,
  });

  kmzLayer.addListener("status_changed", () => {
    if (kmzLayer.getStatus() !== "OK") {
      console.error("Error loading KMZ layer:", kmzLayer.getStatus());
    }
  });

  kmzLayers.push(kmzLayer);
}

function onSearchVessel() {
  document.getElementById("searchButton").addEventListener("click", () => {
    const searchTerm = document.getElementById("vesselSearch").value.trim();
    if (markers.hasOwnProperty(searchTerm)) {
      const marker = markers[searchTerm];
      smoothPanTo(marker.position);
      map.setZoom(20);
      getDataKapalMarker(searchTerm);
    } else {
      alert("Vessel not found.");
    }
  });
}