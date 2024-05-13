function notifikasi(icon, message, title) {
    Swal.fire({
        title: title,
        text: message,
        icon: icon
    });
}

function handleRespon(response) {
    switch (response.status) {
        case "success":
        case "error":
        case "warning":
            notifikasi(response.icon, response.message, response.title);
            break;
        default:
            console.error("Unknown status:", response.status);
    }
}
