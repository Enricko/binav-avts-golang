const websocketUrl = `ws://localhost:8080/ws/kapal`;

function connectWebSocket() {
    websocket = new WebSocket(websocketUrl);

    websocket.onopen = function() {
        console.log('WebSocket connection established.');
    };

    websocket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        for (const device in data) {
            if (data.hasOwnProperty(device)) {
                console.log(device);
                // const gga = data[device].gga;
                // const parsedData = parseGGA(gga);
                // updateMarkerPosition(device, parsedData.latitude, parsedData.longitude);
            }
        }
    };

    websocket.onclose = function(event) {
        console.log('WebSocket connection closed. Reconnecting...');
        setTimeout(connectWebSocket, reconnectInterval);
    };

    websocket.onerror = function(error) {
        console.log('WebSocket error: ', error);
        websocket.close();
    };
}

// Parse NMEA GGA sentence
function parseGGA(gga) {
    const fields = gga.split(',');
    const latitudeDMS = parseFloat(fields[2]);
    const latitudeDirection = fields[3];
    const longitudeDMS = parseFloat(fields[4]);
    const longitudeDirection = fields[5];

    const latitude = convertDMSToDecimal(latitudeDMS, latitudeDirection);
    const longitude = convertDMSToDecimal(longitudeDMS, longitudeDirection);

    // console.log('Latitude:', latitude, 'Longitude:', longitude);
    return { latitude, longitude };
}

function convertDMSToDecimal(degrees, direction) {
    // Extract degrees, minutes, and seconds
    const d = Math.floor(degrees / 100);
    const m = degrees % 100;
    const s = (degrees - (d * 100) - m) * 60;

    // Convert to decimal degrees
    let decimalDegrees = d + (m / 60) + (s / 3600);

    // Adjust for negative direction
    if (direction === 'S' || direction === 'W') {
        decimalDegrees = -decimalDegrees;
    }

    return decimalDegrees;
}

// Update marker position or create new marker
function updateMarkerPosition(device, latitude, longitude) {
    const customMarkerIcon = {
        url: 'https://binav-avts.id/assets/assets/ship.png', // Replace with your custom marker image URL
        scaledSize: calculateMarkerSize(map.getZoom()),
    };

    if (!markers.hasOwnProperty(device)) {
        markers[device] = new google.maps.Marker({
            position: { lat: -1.2692, lng: 116.8253 },
            map: map,
            title: device,
            icon: customMarkerIcon,
        });
    } else {
        markers[device].setPosition({ lat: -1.2692, lng: 116.8253 });
        markers[device].setIcon(customMarkerIcon); // Update marker icon
    }
}

// Update marker sizes based on zoom level
function updateMarkerSizes(zoom) {
    for (const device in markers) {
        if (markers.hasOwnProperty(device)) {
            const marker = markers[device];
            marker.setIcon({
                url: "https://binav-avts.id/assets/assets/ship.png",
                scaledSize: calculateMarkerSize(zoom),
            });
        }
    }
}

// Start WebSocket connection and map initialization
connectWebSocket();