// Global variables
let dataTableMap = {};

// Function to initialize or reload DataTable
function initializeDataTable(modalId, tableId, ajaxUrl, columns) {
  $(`#${modalId}`).on("shown.bs.modal", function () {
    if (dataTableMap[tableId]) {
      // If DataTable instance exists, update its ajax URL and reload
      dataTableMap[tableId].dataTable.ajax.url(ajaxUrl).load();
    } else {
      // If DataTable instance doesn't exist, initialize it
      let dataTable = $(`#${tableId}`).DataTable({
        ajax: ajaxUrl,
        processing: true,
        serverSide: true,
        searching: true,
        responsive: true,
        columnDefs: [
          { orderable: true, targets: 0 },
          { orderable: false, targets: "_all" },
        ],
        columns: columns,
      });

      // Reload the table data every 60 seconds
      let reloadInterval = setInterval(() => {
        dataTable.ajax.reload(null, false);
      }, 60000);

      dataTableMap[tableId] = { dataTable, reloadInterval };
    }
  });

  $(`#${modalId}`).on("hidden.bs.modal", function () {
    if (dataTableMap[tableId]) {
      clearInterval(dataTableMap[tableId].reloadInterval);
    }
  });
}
