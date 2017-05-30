// The response from the server should be an array or objects, 
// that will be passed to the nodeCreator function
//
// The stream starts by calling the init function and passing and id and an option object
// which has the following defaults
// 
// results should have and id field
function Stream(id, opts) {
  if (id == undefined) {
    throw ('Stream id can\'t be blank');
  }

  if (opts == undefined) {
    throw ('Stream options can\' be blank');
  }

  this.id = id;

  if (this.id[0] != "#") {
    this.id = "#" + this.id
  }

  $(this.id).append('<div id="' + this.id.substr(1) + '-content"></div>')
  $(this.id).append('<div id="' + this.id.substr(1) + '-footer">Loading...</div>')

  this._options = {
    URI: "",
    params: {},
    loadCount: 10,
    nodeCreator: function() {},
    beforeLoad: function() {},
    afterLoad: function() {},
    onFail: function() {}
  }
  $.extend(this._options, opts)
  this._setDefaultParams();

  this._freezed = false;
  this._loading = false;
  this._lastLoadCount = 1000;

  var that = this;
  $(window).scroll(function() {
    if ($(that.id + '-footer').visible(true, true)) {
      that._load();
    }
  });

  this._load();
}

Stream.prototype._setDefaultParams = function() {
  $.extend(this._options.params, {
    last_item_id: Infinity,
    page: 0
  })
}

Stream.prototype.freeze = function() {
  this._freezed = true;
}

Stream.prototype.unfreeze = function() {
  this._freezed = false;
}

Stream.prototype._resetLoader = function() {
  $(this.id + '-footer').text("Loading...")
}

Stream.prototype._reset = function() {
  $(this.id + '-content').empty()
  this._lastLoadCount = 1000;
  this._setDefaultParams();
  this._loading = false;
  this._resetLoader();
  this._load();
}

Stream.prototype.changeParams = function(params) {
  $.extend(this._options.params, params)
  this._reset();
}

Stream.prototype.newParams = function(params) {
  this._options.params = params
  this._reset()
}

Stream.prototype._load = function() {
  if (this._loading || this._lastLoadCount < this._options.loadCount || this._freezed) {
    return;
  }
  this._loading = true;
  this._options.beforeLoad();

  var that = this;
  $.post(this._options.URI, this._options.params)
    .done(function(data, status, request) {
      if (!data) {
        return;
      }
      _content = '';
      for (var i = 0; i < data.length; i++) {
        _content = _content + that._options.nodeCreator(data[i])
      }
      $(that.id + '-content').append(_content)

      that._lastLoadCount = data.length;
      $.extend(that._options.params, {
        page: that._options.params.page + 1
      })
      if (data.length > 0) {
        $.extend(that._options.params, {
          last_item_id: data[data.length - 1].id || Infinity
        })
      }

      that._options.afterLoad();

      if ($(that.id + '-footer').visible(true, true)) {
        that._loading = false;
        that._load();
      }
    })
    .fail(function(response) {
      that._options.onFail(response);
    })
    .always(function() {
      that._loading = false;
      if (that._lastLoadCount < that._options.loadCount) {
        $(that.id + '-footer').text("No more items :(")
      }
    });
}


// TODOs: loader styling
// handle different sorting and cases of freezing stream
