var notifications = {};
notifications.unread = 0;
notifications.initialized = false;
notifications.topId = 0;
notifications.lastId = 0;
notifications.lastLoadCount = 1000;
notifications.loading = false;
notifications._loadCount = 3;

notifications.loadNew = function() {
  // if (!notifications.initialized) {
  //   return;
  // }
  $.post("/notifications/new", {
    top_id: notifications.topId
  })
    .done(function(data, status) {
      if (!data) {
        return;
      }

      // always set topId
      if (data.length > 0) {
        notifications.topId = data[0].id;
      }

      // set lastId if there are no notifications
      if (data.length > 0 && notifications.isEmpty()) {
        notifications.lastId = data[data.length - 1].id;
      }

      var n = '';
      for (var i = 0; i < data.length; i++) {
        n = n + notifications.getNotificationNode(data[i])
      }
      $("#notifications").prepend(n)
    });
}

notifications.load = function() {
  if (notifications.loading || notifications.lastLoadCount < notifications._loadCount) {
    return;
  }
  notifications.loading = true;
  $.post("/notifications", {
    last_id: notifications.lastId
  })
    .done(function(data, status) {
      if (!data) {
        return;
      }

      // always set topId
      if (data.length > 0) {
        notifications.lastId = data[data.length - 1].id;
      }

      // set topId if there are no notifications
      if (data.length > 0 && notifications.isEmpty()) {
        notifications.topId = data[0].id;
      }

      notifications.lastLoadCount = data.length;

      var n = '';
      for (var i = 0; i < data.length; i++) {
        n = n + notifications.getNotificationNode(data[i])
      }
      $("#notifications").append(n)
    })
    .always(function() {
      notifications.loading = false;
    });
}

notifications.isEmpty = function() {
  return $("#notifications")[0].innerHTML.trim(" ") == "";
}

notifications.checkUnread = function() {
  $.post("/notifications/unread")
    .done(function(data, status) {
      if (!data) {
        return;
      }
      if (data.unread > 0) {
        notifications.updateUnread(data.unread)
      }
      notifications.unread = data.unread;
    })
    .always(function() {
      notifications.initialized = true;
    });
}

notifications.updateUnread = function(unread) {
  $("#notifications-counter")[0].innerHTML = unread;
}

notifications.getNotificationNode = function(notification) {
  var table = notification.table_name.replace(/_/gi, ' ');
  var hint = '';
  if (notification.type == "c") {
    hint = `<div class="_hint">comment on your post on <span class="_table">` + table + `</span></div>`
  }
  if (notification.type == "r") {
    hint = `<div class="_hint">reply to your comment on <span class="_table">` + table + `</span></div>`
  }

  var readClass = "";
  if (notification.read) {
    readClass = "_read"
  }
  return `<div class="notification ` + readClass + `" id="notification-` + notification.id + `" onclick="notifications.read(` + notification.id + `)"><div class="_ctrl"><span onclick="notifications.read(` + notification.id + `)">R</span><span>X</span></div>` + hint + `<div class="_content"><div class="_title">` + notification.title + `</div><div class="_body">` + notification.body + `</div></div></div>`
}

notifications.read = function(id) {
  if (!$("#notification-" + id)[0].classList.contains("_read")) {
    $("#notification-" + id).addClass("_read")
    $.post("/notifications/read", {
      notification_id: id
    })
      .done(function(data, status) {});
  }
}

$(function() {
  // notifications.checkUnread();

  $('#notifications-dropdown').on('show.bs.dropdown', function() {
    if (notifications.isEmpty()) {
      notifications.loadNew();
    // return;
    }

  // if (notifications.unread == 0 && notifications.isEmpty()) {
  //   notifications.load();
  // }
  })


  $('#notifications').scrollbar({
    disableBodyScroll: true,
    onScroll: function(y, x) {
      if (notifications.isEmpty()) {
        return;
      }
      if (y.scroll == y.maxScroll) {
        notifications.load();
      }
    }
  });

})