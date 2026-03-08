// ─── Quantity controls ─────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', function () {
  // Qty +/- buttons
  document.querySelectorAll('.qty-control').forEach(function (ctrl) {
    var minus = ctrl.querySelector('.qty-minus');
    var plus  = ctrl.querySelector('.qty-plus');
    var input = ctrl.querySelector('.qty-input');
    if (!minus || !plus || !input) return;

    function getStep() {
      return parseFloat(input.step) || 1;
    }
    function getVal() {
      return parseFloat(input.value) || 0;
    }

    minus.addEventListener('click', function () {
      var v = getVal() - getStep();
      input.value = Math.max(0, parseFloat(v.toFixed(2)));
      input.dispatchEvent(new Event('input'));
    });

    plus.addEventListener('click', function () {
      var v = getVal() + getStep();
      input.value = parseFloat(v.toFixed(2));
      input.dispatchEvent(new Event('input'));
    });
  });

  // Highlight cards with qty > 0 and update footer summary
  var form = document.getElementById('order-form');
  var footer = document.getElementById('order-footer');
  var summary = document.getElementById('order-summary');

  function updateSummary() {
    if (!form || !summary) return;
    var inputs = form.querySelectorAll('.qty-input');
    var count = 0;
    inputs.forEach(function (inp) {
      var v = parseFloat(inp.value) || 0;
      var card = inp.closest('.order-card');
      if (v > 0) {
        count++;
        if (card) card.classList.add('has-qty');
      } else {
        if (card) card.classList.remove('has-qty');
      }
    });
    summary.textContent = count + ' article' + (count > 1 ? 's' : '') + ' sélectionné' + (count > 1 ? 's' : '');
  }

  if (form) {
    form.querySelectorAll('.qty-input').forEach(function (inp) {
      inp.addEventListener('input', updateSummary);
    });
    updateSummary();
  }

  // ─── Image preview on article form ────────────────────────────────────────
  var fileInput = document.getElementById('image');
  var preview   = document.getElementById('image-preview');
  var fileText  = document.getElementById('file-upload-text');

  if (fileInput && preview) {
    fileInput.addEventListener('change', function () {
      var file = fileInput.files[0];
      if (!file) return;
      var reader = new FileReader();
      reader.onload = function (e) {
        preview.src = e.target.result;
        preview.style.display = 'block';
        if (fileText) fileText.textContent = file.name;
      };
      reader.readAsDataURL(file);
    });
  }

  // Drag-over highlight on file upload area
  var uploadArea = document.getElementById('file-upload-area');
  if (uploadArea) {
    uploadArea.addEventListener('dragover', function (e) {
      e.preventDefault();
      uploadArea.classList.add('drag-over');
    });
    uploadArea.addEventListener('dragleave', function () {
      uploadArea.classList.remove('drag-over');
    });
    uploadArea.addEventListener('drop', function (e) {
      e.preventDefault();
      uploadArea.classList.remove('drag-over');
      if (fileInput && e.dataTransfer.files.length) {
        fileInput.files = e.dataTransfer.files;
        fileInput.dispatchEvent(new Event('change'));
      }
    });
  }
});
