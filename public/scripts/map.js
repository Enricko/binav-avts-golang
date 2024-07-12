let map;
  let vesselOverlay;
  const kmzLayers = [];

  const initialZoom = 16; // Initial zoom level

  let isRulerOn = false;

  function createRulerButton(map) {
    const controlButton = document.createElement("button");

    controlButton.classList.add("btn", "btn-primary", "rounded-circle", "ml-4");
    controlButton.style.backgroundColor = "white";
    controlButton.style.border = "0";
    controlButton.style.width = "50px";
    controlButton.style.height = "50px";
    controlButton.style.marginRight = "0.5rem";
    controlButton.style.backgroundColor = "white";
    controlButton.innerHTML =
      '<i class="fas fa-solid fa-ruler" style="color: black;"></i>';
    controlButton.type = "button";
    controlButton.title = "Ruler Button";
    controlButton.addEventListener("click", () => {
      isRulerOn = !isRulerOn;
      initMap();
    });
    return controlButton;
  }

  function initMap() {
    map = new google.maps.Map(document.getElementById("map"), {
      zoom: initialZoom,
      center: { lat: -1.0574568371666666, lng: 117.35121627983332 }, // Replace with the initial coordinates

      disableDefaultUI: true,
      zoomControl: true,
      scaleControl: true,
    });

    /// CREATE RULER BUTTON
    const centerControlDiv = document.createElement("div");
    // Create the control.
    const centerControl = createRulerButton(map);
    // Append the control to the DIV.
    centerControlDiv.appendChild(centerControl);
    map.controls[google.maps.ControlPosition.RIGHT_BOTTOM].push(
      centerControlDiv
    );

    /// POLYLINE
    poly = new google.maps.Polyline({
      strokeColor: "#000000",
      strokeOpacity: 1.0,
      strokeWeight: 3,
    });
    poly.setMap(map);
    // Add a listener for the click event
    if (isRulerOn) {
      map.addListener("click", addLatLng);
    }

    fetchKMZFiles();
    connectWebSocket();

    document
      .getElementById("searchButton")
      .addEventListener("click", function () {
        const searchTerm = document.getElementById("vesselSearch").value.trim();
        console.log("Search term:", searchTerm);
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
  // TODO : KALAU GK GUNA HAPUS AJA
  // function addLatLng(event) {
  //   const path = poly.getPath();

  //   // Because path is an MVCArray, we can simply append a new coordinate
  //   // and it will automatically appear.
  //   path.push(event.latLng);
  //   // Add a new marker at the new plotted point on the polyline.
  //   new google.maps.Marker({
  //     position: event.latLng,
  //     title: "#" + path.getLength(),
  //     map: map,
  //   });
  // }

  function smoothPanTo(latLng) {
    const panSteps = 30; // Number of steps for the pan animation
    const panDuration = 1000; // Duration of the pan animation in milliseconds
    const panInterval = panDuration / panSteps; // Interval between each step

    const startLat = map.getCenter().lat();
    const startLng = map.getCenter().lng();
    const endLat = latLng.lat;
    const endLng = latLng.lng;

    let step = 0;

    function panStep() {
      const lat = startLat + (endLat - startLat) * (step / panSteps);
      const lng = startLng + (endLng - startLng) * (step / panSteps);
      map.setCenter({ lat: lat, lng: lng });

      if (step < panSteps) {
        step++;
        setTimeout(panStep, panInterval);
      }
    }

    panStep();
  }

  // url: 'https://maps.google.com/mapfiles/ms/icons/red-dot.png',
  function fetchKMZFiles() {
    fetch("/mappings")
      .then((response) => response.json())
      .then((data) => {
        data.forEach((mapping) => {
          if (mapping.status) {
            loadKMZLayer(mapping.file);
          }
        });
      })
      .catch((error) => console.error("Error fetching mappings:", error));
  }

  function loadKMZLayer(filePath) {
    console.log("Loading KMZ layer:", filePath);
    // const kmzUrl = `${getBaseURL()}public/${filePath}`;
    const kmzUrl = `https://golang.binav-avts.id/public/${filePath}`;
    // const kmzUrl = `https://a4d6-140-213-164-88.ngrok-free.app/public/testMWPA.kmz`;
    const kmzLayer = new google.maps.KmlLayer({
      url: kmzUrl,
      map: map,
      preserveViewport: true, // Prevent the map from zooming to fit the KML/KMZ
    });
    kmzLayer.addListener("status_changed", function () {
      if (kmzLayer.getStatus() !== "OK") {
        console.error("Error loading KMZ layer:", kmzLayer.getStatus());
      }
    });

    kmzLayers.push(kmzLayer);
  }

  