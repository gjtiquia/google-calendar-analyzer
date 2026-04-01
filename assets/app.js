(function () {
  "use strict";

  /**
   * @param {number} utcMs
   * @param {string} timeZone
   * @returns {{ year: number, month: number, day: number, hour: number, minute: number, second: number }}
   */
  function formatPartsInTZ(utcMs, timeZone) {
    var parts = new Intl.DateTimeFormat("en-CA", {
      timeZone: timeZone,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }).formatToParts(new Date(utcMs));
    function get(type) {
      var p = parts.find(function (x) {
        return x.type === type;
      });
      return p ? parseInt(p.value, 10) : 0;
    }
    return {
      year: get("year"),
      month: get("month"),
      day: get("day"),
      hour: get("hour"),
      minute: get("minute"),
      second: get("second"),
    };
  }

  /**
   * First instant of calendar dateStr (YYYY-MM-DD) in timeZone, as RFC3339 UTC.
   * @param {string} dateStr
   * @param {string} timeZone
   * @returns {string}
   */
  function zonedDateStartToRFC3339UTC(dateStr, timeZone) {
    var parts = dateStr.split("-");
    var y = parseInt(parts[0], 10);
    var mo = parseInt(parts[1], 10);
    var da = parseInt(parts[2], 10);
    if (typeof Temporal !== "undefined") {
      try {
        var plain = Temporal.PlainDate.from({ year: y, month: mo, day: da });
        var zdt = plain.toZonedDateTime({
          timeZone: timeZone,
          plainTime: Temporal.PlainTime.from("00:00:00"),
        });
        return zdt.toInstant().toString({ smallestUnit: "second" });
      } catch (e) {
        /* fall through */
      }
    }
    var start = Date.UTC(y, mo - 1, da - 3, 0, 0, 0);
    var end = Date.UTC(y, mo - 1, da + 3, 0, 0, 0);
    for (var t = start; t <= end; t += 60000) {
      var p = formatPartsInTZ(t, timeZone);
      if (p.year === y && p.month === mo && p.day === da) {
        return new Date(t).toISOString().replace(/\.\d{3}Z$/, "Z");
      }
    }
    return new Date(Date.UTC(y, mo - 1, da, 0, 0, 0))
      .toISOString()
      .replace(/\.\d{3}Z$/, "Z");
  }

  /**
   * @param {HTMLFormElement} form
   * @returns {URLSearchParams}
   */
  function buildExportParams(form) {
    var tzEl = document.getElementById("tz");
    var startEl = document.getElementById("start");
    var endEl = document.getElementById("end");
    var tz = tzEl && tzEl.value ? tzEl.value : "UTC";
    var startVal = startEl && startEl.value;
    var endVal = endEl && endEl.value;
    var params = new URLSearchParams(new FormData(form));
    if (startVal && endVal) {
      params.set("start", zonedDateStartToRFC3339UTC(startVal, tz));
      params.set("end", zonedDateStartToRFC3339UTC(endVal, tz));
    }
    params.set("tz", tz);
    return params;
  }

  function localYMD(d) {
    var y = d.getFullYear();
    var m = String(d.getMonth() + 1).padStart(2, "0");
    var day = String(d.getDate()).padStart(2, "0");
    return y + "-" + m + "-" + day;
  }

  function initQueryForm() {
    var form = document.getElementById("query-form");
    var tzSelect = document.getElementById("tz");
    var startInput = document.getElementById("start");
    var endInput = document.getElementById("end");
    if (!form || !tzSelect || !startInput || !endInput) return;

    var detected =
      (typeof Intl !== "undefined" &&
        Intl.DateTimeFormat &&
        Intl.DateTimeFormat().resolvedOptions().timeZone) ||
      "UTC";

    tzSelect.innerHTML = "";

    var zones = [];
    if (typeof Intl !== "undefined" && typeof Intl.supportedValuesOf === "function") {
      try {
        zones = Intl.supportedValuesOf("timeZone");
      } catch (e) {
        zones = [];
      }
    }
    if (zones.length === 0) {
      zones = ["UTC", detected];
    }

    var seen = {};
    for (var i = 0; i < zones.length; i++) {
      var z = zones[i];
      if (seen[z]) continue;
      seen[z] = true;
      var opt = document.createElement("option");
      opt.value = z;
      opt.textContent = z;
      tzSelect.appendChild(opt);
    }

    if (seen[detected]) {
      tzSelect.value = detected;
    } else {
      var opt2 = document.createElement("option");
      opt2.value = detected;
      opt2.textContent = detected;
      tzSelect.insertBefore(opt2, tzSelect.firstChild);
      tzSelect.value = detected;
    }

    var dStart = new Date();
    dStart.setDate(dStart.getDate() - 1);
    var dEnd = new Date();
    dEnd.setDate(dEnd.getDate() + 7);
    startInput.value = localYMD(dStart);
    endInput.value = localYMD(dEnd);

    document.body.addEventListener("htmx:configRequest", function (evt) {
      var detail = evt.detail;
      if (!detail || !detail.parameters) return;
      var path = detail.path || "";
      if (path.indexOf("/events/query") === -1) return;
      var tzEl = document.getElementById("tz");
      var startEl = document.getElementById("start");
      var endEl = document.getElementById("end");
      if (!tzEl || !startEl || !endEl) return;
      var tz = tzEl.value || "UTC";
      var sv = startEl.value;
      var ev = endEl.value;
      if (!sv || !ev) return;
      detail.parameters.set("start", zonedDateStartToRFC3339UTC(sv, tz));
      detail.parameters.set("end", zonedDateStartToRFC3339UTC(ev, tz));
      detail.parameters.set("tz", tz);
    });

    var exportBtn = document.getElementById("export-csv-btn");
    if (exportBtn) {
      exportBtn.addEventListener("click", function () {
        var params = buildExportParams(form);
        window.location = "/events/export.csv?" + params.toString();
      });
    }
  }

  document.addEventListener("DOMContentLoaded", initQueryForm);
})();
