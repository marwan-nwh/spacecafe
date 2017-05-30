// https://developer.mozilla.org/en-US/docs/Web/API/document.cookie

// http://stackoverflow.com/a/14525299/1420619
var paramterize = function(data) {
  return Object.keys(data).map(function(k) {
    return encodeURIComponent(k) + '=' + encodeURIComponent(data[k])
  }).join('&')
};

const APIPath = ""

// HTTP
// TODO: handel 400, 401, 404 and 500
var act = function(xhr, success, error) {
  error = error || function() {};
  return function() {
    if (xhr.readyState == 4) {
      var response = (xhr.response == "") ? xhr.response : JSON.parse(xhr.response);
      if (xhr.status == 200) {
        success(response, xhr.status);
      } else {
        error(response, xhr.status);
      }
    }
    ;
  }
}

var http = {
  get: function(url, success, error, timeout, ontimeout) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = act(xhr, success, error);
    xhr.open("GET", APIPath + url, true);
    xhr.setRequestHeader("Token", cookies.get('token'));
    xhr.timeout = timeout;
    xhr.ontimeout = ontimeout;
    xhr.send();
  },
  post: function(url, params, success, error, timeout, ontimeout) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = act(xhr, success, error);
    xhr.open("POST", APIPath + url, true);
    xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    xhr.setRequestHeader("Token", cookies.get('token'));
    xhr.timeout = timeout;
    xhr.ontimeout = ontimeout;
    xhr.send(paramterize(params));

  },
  upload: function(url, params, success, error, timeout, ontimeout) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = act(xhr, success, error);
    xhr.open("POST", APIPath + url, true);
    xhr.setRequestHeader("Token", cookies.get('token'));
    xhr.timeout = timeout;
    xhr.ontimeout = ontimeout;
    var formData = new FormData();
    for (i in params) {
      formData.append(i, params[i]);
    }
    xhr.send(formData);
  }
}