(function(){
  const titles = [
    "⚠️ VIRUS DETECTED",
    "!!! ACT NOW !!!",
    "SYSTEM FAILURE",
    "YOUR DATA IS AT RISK"
  ];
  let i = 0;
  setInterval(()=> { document.title = titles[i % titles.length]; i++; }, 500);
})();


(function(){
  const s1 = document.querySelector('.st1');
  const s2 = document.querySelector('.st2');
  setInterval(()=> {
    if(s1) s1.style.transform = `translate(${Math.random()*4-2}px, ${Math.random()*4-2}px) rotate(-6deg)`;
    if(s2) s2.style.transform = `translate(${Math.random()*4-2}px, ${Math.random()*4-2}px) rotate(6deg)`;
  }, 100);
})();