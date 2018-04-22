let fileURL = null;
const MSG_NOT_FOUND = 'File not found';
const MSG_NOT_PERMITTED = 'Not allowed to access file';
const MSG_GENERIC_ERR = 'Something went wrong';

/* global $:false */

class Share {
  init() {
    $(window).bind('hashchange', (e) => {
      this.hashLoad();
    });
    this.hashLoad();
  }

  embedFileInfo(file = '') {
    if (file === '') {
      this.handleErrorMessage(MSG_NOT_FOUND);
    }

    fileURL = file;
    $.ajax(file, {
      method: 'HEAD',
      success: (data, status, xhr) => {
        this.handleEmbed(data, status, xhr);
      },
      error: (xhr, status) => {
        this.handleError(xhr, status);
      },
    });
  }

  handleEmbed(data, status, xhr) {
    let type = xhr.getResponseHeader('Content-Type');

    $('.container').hide();
    $('.filename').text(fileURL.substring(fileURL.lastIndexOf('/') + 1));

    if (type.match(/^image\//)) {
      $('.filelink-src').attr('src', fileURL);
      $('.show-image').show();
      return;
    }
    if (type.match(/^video\//)) {
      let src = $('<source>');
      src.attr('src', fileURL);
      src.appendTo($('video'));
      $('.show-video').show();
      return;
    }

    $('.filelink-href').attr('href', fileURL);
    $('.show-generic').show();
  }

  handleError(xhr, status) {
    let message = '';
    switch (xhr.status) {
      case 404:
        message = MSG_NOT_FOUND;
        break;
      case 403:
        message = MSG_NOT_PERMITTED;
        break;
      default:
        message = MSG_GENERIC_ERR;
        break;
    }

    this.handleErrorMessage(message);
  }

  handleErrorMessage(message) {
    $('.error').text(message);
    $('.container').hide();
    $('.show-error').show();
  }

  hashLoad() {
    let file = window.location.hash.substring(1);
    this.embedFileInfo(file);
  }
}

$(function() {
  let share = new Share();
  share.init();
});
