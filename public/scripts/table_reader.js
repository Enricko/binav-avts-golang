// Global variables
let dataTableMap = {};

// Function to initialize or reload DataTable
function initializeDataTable(modalId, tableId, ajaxUrl, columns) {
  $(`#${modalId}`).on("shown.bs.modal", function () {
    let dataTable;

    if ($.fn.DataTable.isDataTable(`#${tableId}`)) {
      dataTable = $(`#${tableId}`).DataTable();
      dataTable.ajax.reload(null, false);
    } else {
      dataTable = $(`#${tableId}`).DataTable({
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
    }

    // Reload the table data every 60 seconds
    let reloadInterval = setInterval(() => {
      dataTable.ajax.reload(null, false);
    }, 60000);

    dataTableMap[tableId] = { dataTable, reloadInterval };
  });

  $(`#${modalId}`).on("hidden.bs.modal", function () {
    clearInterval(dataTableMap[tableId].reloadInterval);
  });
}
