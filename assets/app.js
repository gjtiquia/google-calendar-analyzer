document.addEventListener("DOMContentLoaded", function () {
  var btn = document.getElementById("export-csv-btn");
  var form = document.getElementById("query-form");
  if (!btn || !form) return;
  btn.addEventListener("click", function () {
    var params = new URLSearchParams(new FormData(form));
    window.location = "/events/export.csv?" + params.toString();
  });
});
