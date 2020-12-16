var shogiboard = shogiboard || {};
(function (global) {
  var _ = shogiboard;

  const W = 43;
  const H = 48;
  const GRID_COLOR = "#000000";
  
  const KOMA_MATRIX = [
    [null, "R", "B", "G", "S", "N", "L", "P"],
    ["K", "+R", "+B", "", "+S", "+N", "+L", "+P"],
    null,
    [null, "r", "b", "g", "s", "n", "l", "p"],
    ["k", "+r", "+b", null, "+s", "+n", "+l", "+p"],
    null,
  ];
  
  var KOMA = {};
  for (let i = 0; i < 6; i++) {
    const row = KOMA_MATRIX[i];
    if (row === null) {
      continue;
    }
  
    for (let j = 0; j < 8; j++) {
      const cell = row[j];
      if (cell === null) {
        continue;
      }
  
      KOMA[cell] = {
        sx: j * W,
        sy: i * H,
        b: i < 3,
      };
    }
  }
  
  const NUMBER = " 123456789";
  
  function onloadDraw(ctx, img, sfen) {
    for (let i = 0; i <= 11; i++) {
      for (let j = 0; j <= 11; j++) {
        ctx.drawImage(img, KOMA[""].sx, KOMA[""].sy, W, H, i*W, j*H, W, H);
      }
    }
  
    const fs = sfen.split(' ');
    if (fs.length != 4) {
      return;
    }
  
    const [b, p, c, q] = fs;
    
    b.split('/').forEach((row, i) => {
      var j = 0;
      var sym = "";
      for (let k = 0; k < row.length; k++) {
        let idx = NUMBER.indexOf(row[k]);
        if (idx !== -1) {
          j = j + idx;
          continue;
        }
  
        sym = sym + row[k];
        if (sym === "+") {
          continue;
        }
  
        src = KOMA[sym];
        ctx.drawImage(img, src.sx, src.sy, W, H, (j+1)*W, (i+1)*H, W, H);
        j++;
        sym = "";
      }
    });
  
    ctx.font = "20px serif";
    var bidx = 0;
    var widx = 0;
    var prev = null;
    var num = "";
    c.split('').forEach((k) => {
      let n = NUMBER.indexOf(k);
      if (n !== -1) {
        num = num + k;
        return;
      }
  
      koma = KOMA[k];
      if (koma.b) {
        ctx.drawImage(img, koma.sx, koma.sy, W, H, (bidx+1)*W, 10*H, W, H);
        if (num !== "") {
          ctx.fillText(num, (bidx+1.7)*W, 11*H);
        }
        bidx++;
      } else {
        ctx.drawImage(img, koma.sx, koma.sy, W, H, (widx+1)*W, 0, W, H);
        if (num !== "") {
          ctx.fillText(num, (widx+1.7)*W, H);
        }
        widx++;
      }
      num = "";
      prev = koma;
    });
  
    ctx.beginPath();
    ctx.strokeStyle = GRID_COLOR;
  
    for (let i = 1; i <= 10; i++) {
      ctx.moveTo(W, H*i);
      ctx.lineTo(W*10, H*i);
    }
  
    for (let i = 1; i <= 10; i++) {
      ctx.moveTo(W*i, H);
      ctx.lineTo(W*i, H*10);
    }
  
    for (let i = 1; i <= 2; i++) {
      for (let j = 1; j <= 2; j++) {
        let x = (1 + 3*i) * W;
        let y = (1 + 3*j) * H;
        ctx.moveTo(x, y);
        ctx.arc(x, y, 2, 0, 2*Math.PI, false);
        ctx.fill();
      }
    }
  
    ctx.stroke();
  }
  
  _.draw = function(id, komafile, sfen) {
    const canvas = document.getElementById(id);
    canvas.width = W * 11;
    canvas.height = H * 11;
  
    const ctx = canvas.getContext('2d');
  
    const img = new Image();
    img.onload = function(e) {
      onloadDraw(ctx, img, sfen);
    }
    img.src = komafile + '?' + new Date().getTime();
  
    return ctx;
  };
}(this));
