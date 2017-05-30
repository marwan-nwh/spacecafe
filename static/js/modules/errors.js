var Errors = {};

Errors.handle = function(error) {
  const timestamp = Math.floor(Date.now() / 1000);
  if (error.status == 406) { // alert using response message
    Errors.showAlert(timestamp, error.responseJSON.message);
    Errors.fadeAlert(timestamp);
  }
}

Errors.fadeAlert = function(id) {
  window.setTimeout(function() {
    if ($("#alert-id")[0].innerHTML != id.toString()) {
      return;
    }
    $("#alert").fadeTo(800, 0).slideUp(800, function() {
      Errors.hideAlert();
    });
  }, 4000);
}

Errors.showAlert = function(id, msg) {
  $("#alert").removeAttr("style");
  $("#alert-text").text(msg);
  $("#alert-id").text(id)
  $("#alert").removeClass("hide");
}

Errors.hideAlert = function() {
  $("#alert").addClass("hide");
}