class BaseVesselOverlay extends google.maps.OverlayView {
  constructor(map, position, top, left, width, height, rotationAngle, imageMap) {
    super();
    this.map = map;
    this.position = position;
    this.offsetFromCenter = { x: left, y: top };
    this.vesselDimensions = { width, height };
    this.rotationAngle = rotationAngle;
    this.imageMap = imageMap;
    this.div = null;
    this.setMap(map);
  }

  onAdd() {
    this.div = document.createElement("div");
    this.div.style.position = "absolute";
    
    const img = document.createElement("img");
    img.style.width = "100%";
    img.style.height = "100%";
    img.style.transformOrigin = `${(this.offsetFromCenter.x / this.vesselDimensions.width) * 100}% ${(this.offsetFromCenter.y / this.vesselDimensions.height) * 100}%`;
    img.style.transform = `rotate(${this.rotationAngle}deg)`;
    
    this.div.appendChild(img);
    this.getPanes().overlayMouseTarget.appendChild(this.div);
    
    this.updateImage();
  }

  draw() {
    const projection = this.getProjection();
    if (!projection) return;

    const position = projection.fromLatLngToDivPixel(this.position);
    const scale = this.getScale();
    
    const scaledWidth = this.vesselDimensions.width * scale;
    const scaledHeight = this.vesselDimensions.height * scale;
    const offsetX = this.metersToPixels(this.offsetFromCenter.x, this.position.lat, scale);
    const offsetY = this.metersToPixels(this.offsetFromCenter.y, this.position.lat, scale);

    if (this.div) {
      Object.assign(this.div.style, {
        left: `${position.x - offsetX}px`,
        top: `${position.y - offsetY}px`,
        width: `${scaledWidth}px`,
        height: `${scaledHeight}px`,
        zIndex: '999'
      });
    }

    // if (this.nameLabel) {
    //   const scale = this.getScale();
    //   this.nameLabel.style.fontSize = `${Math.max(8 * scale, 6)}px`;  // Minimum font size of 8px
    //   this.nameLabel.style.bottom = `-${Math.max(13 * scale, 10)}px`;  // Adjust bottom margin
    // }
  }

  getScale() {
    return 1 / ((156543.03392 * Math.cos((this.position.lat * Math.PI) / 180)) / Math.pow(2, this.map.getZoom()));
  }

  metersToPixels(meters, latitude, scale) {
    return meters / ((156543.03392 * Math.cos((latitude * Math.PI) / 180)) / Math.pow(2, this.map.getZoom()));
  }

  updateImage() {
    // Implement in child classes
  }

  onRemove() {
    if (this.div) {
      this.div.parentNode.removeChild(this.div);
      this.div = null;
    }
  }
}

class VesselOverlay extends BaseVesselOverlay {
  constructor(map, device, position, top, left, width, height, rotationAngle, imageMap, infoContent, status) {
    super(map, position, top, left, width, height, rotationAngle, imageMap);
    this.device = device;
    this.infoContent = infoContent;
    this.status = status;
    this.infoWindow = null;
    this.animationInProgress = false;
  }

  onAdd() {
    super.onAdd();
    if (this.div) {
      this.div.addEventListener("mouseover", () => this.showInfoWindow());
      this.div.addEventListener("mouseout", () => this.hideInfoWindow());
      this.div.addEventListener("dblclick", (e) => this.onDblClick(e));
    }
  }

  updateImage() {
    const color = this.status === "Connected" ? [40, 167, 69] : [220, 53, 69];
    changeImageColor(`/public/upload/assets/image/vessel_map/${this.imageMap}`, color, (dataUrl) => {
      if (dataUrl && this.div && this.div.firstChild) {
        this.div.firstChild.src = dataUrl;
      }
    });
  }

  update(device, newPosition, top, left, width, height, newRotationAngle, imageMap, infoContent, status) {
    if (this.animationInProgress) {
      cancelAnimationFrame(this.animationFrame);
    }

    const oldPosition = this.position;
    const oldRotationAngle = this.rotationAngle;
    this.device = device;
    this.infoContent = infoContent;
    this.status = status;
    this.offsetFromCenter = { x: left, y: top };
    this.vesselDimensions = { width, height };
    this.imageMap = imageMap;

    this.updateImage();
    this.animateMovementAndRotation(oldPosition, newPosition, oldRotationAngle, newRotationAngle);
  }

  animateMovementAndRotation(startPosition, endPosition, startAngle, endAngle) {
    const startTime = performance.now();
    const duration = 1000; // Animation duration in milliseconds

    const animate = (currentTime) => {
      const elapsed = currentTime - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Interpolate position
      const currentLat = startPosition.lat + (endPosition.lat - startPosition.lat) * progress;
      const currentLng = startPosition.lng + (endPosition.lng - startPosition.lng) * progress;
      this.position = {lat :currentLat, lng: currentLng};

      // Interpolate rotation
      this.rotationAngle = startAngle + (endAngle - startAngle) * progress;

      this.draw();
      this.updateRotation();

      if (progress < 1) {
        this.animationInProgress = true;
        this.animationFrame = requestAnimationFrame(animate);
      } else {
        this.animationInProgress = false;
      }
    };

    this.animationFrame = requestAnimationFrame(animate);
  }

  updateRotation() {
    if (this.div && this.div.firstChild) {
      this.div.firstChild.style.transform = `rotate(${this.rotationAngle}deg)`;
    }
  }

  showInfoWindow() {
    if (!this.infoWindow) {
      this.infoWindow = document.createElement("div");
      Object.assign(this.infoWindow.style, {
        position: "absolute",
        backgroundColor: "white",
        border: "1px solid black",
        padding: "5px",
        zIndex: "1000"
      });
      this.infoWindow.innerHTML = this.infoContent;
      this.getPanes().floatPane.appendChild(this.infoWindow);
    }
    const projection = this.getProjection();
    if (projection) {
      const position = projection.fromLatLngToDivPixel(this.position);
      Object.assign(this.infoWindow.style, {
        left: `${position.x}px`,
        top: `${position.y - 50}px`
      });
    }
  }

  hideInfoWindow() {
    if (this.infoWindow && this.infoWindow.parentNode) {
      this.infoWindow.parentNode.removeChild(this.infoWindow);
      this.infoWindow = null;
    }
  }

  onDblClick(event) {
    event.preventDefault();
    if (typeof getDataKapalMarker === 'function') {
      getDataKapalMarker(this.device);
    }
    this.map.setCenter(this.position);
  }
}


class VesselOverlayHistory extends BaseVesselOverlay {
  constructor(map, position, top, left, width, height, rotationAngle, imageMap) {
    super(map, position, top, left, width, height, rotationAngle, imageMap);
    this.img = null;
  }

  onAdd() {
    super.onAdd();
    this.img = this.div.firstChild;
    this.updateRotation();
  }

  updateImage() {
    changeImageColor(`/public/upload/assets/image/vessel_map/${this.imageMap}`, [150, 112, 0], (dataUrl) => {
      if (dataUrl && this.img) {
        this.img.src = dataUrl;
        this.updateRotation();
      }
    });
  }

  updateRotation() {
    if (this.img) {
      this.img.style.transformOrigin = `${(this.offsetFromCenter.x / this.vesselDimensions.width) * 100}% ${(this.offsetFromCenter.y / this.vesselDimensions.height) * 100}%`;
      this.img.style.transform = `rotate(${this.rotationAngle}deg)`;
    }
  }

  update(position, top, left, width, height, rotationAngle, imageMap) {
    this.position = position;
    this.offsetFromCenter = { x: left, y: top };
    this.vesselDimensions = { width, height };
    this.rotationAngle = rotationAngle;
    this.imageMap = imageMap;

    this.updateImage();
    this.updateRotation();
    this.draw();
  }
}

function changeImageColor(imageUrl, color, callback) {
  const img = new Image();
  img.crossOrigin = "Anonymous";
  img.onload = function() {
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    canvas.width = img.width;
    canvas.height = img.height;
    
    ctx.drawImage(img, 0, 0);
    ctx.shadowColor = "rgba(0, 0, 0, 0.5)";
    ctx.shadowBlur = 10;
    ctx.drawImage(img, 0, 0);
    
    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
    const data = imageData.data;
    for (let i = 0; i < data.length; i += 4) {
      if (data[i + 3] !== 0) {
        data[i] = color[0];
        data[i + 1] = color[1];
        data[i + 2] = color[2];
      }
    }
    ctx.putImageData(imageData, 0, 0);
    callback(canvas.toDataURL());
  };
  img.onerror = () => {
    console.error("Failed to load image:", imageUrl);
    callback(null);
  };
  img.src = imageUrl;
}