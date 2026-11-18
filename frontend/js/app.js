import {defaults,validate,timecode,clipLabel} from "./project.js";
const $=id=>document.getElementById(id);
let project=defaults(),url=null;
const video=$("video"), stage=$("stage");
const fields=["format","background","scale","x","y"];
function notify(text){$("status").textContent=text}
function render(){
 stage.className="stage ratio-"+project.format.replace(":","-");
 stage.style.background=project.background;
 $("background-label").textContent=project.background;
 for(const key of fields)$(key).value=project[key];
 for(const key of ["scale","x","y"])$(key+"-value").textContent=project[key]+(key==="scale"?"%":"%");
 $("safe-toggle").checked=project.safe;
 $("safe").classList.toggle("hidden",!project.safe);
 video.style.transform="translate("+project.x+"%, "+project.y+"%) scale("+(project.scale/100)+")";
 $("clipname").textContent=project.media?.name||"Geen media";
 $("clip").textContent=project.media?clipLabel(project.media.name):"Video-track · importeer een opname";
 $("clip").classList.toggle("empty",!project.media);
 $("duration").textContent=timecode(video.duration);
}
function setMedia(file){
 if(!file||!file.type.startsWith("video/")){notify("Kies een videobestand");return}
 if(url)URL.revokeObjectURL(url);
 url=URL.createObjectURL(file);
 video.src=url;video.style.display="block";$("placeholder").hidden=true;
 project.media={name:file.name,size:file.size,type:file.type};
 $("scrubber").disabled=false;
 notify("Media geïmporteerd");
 render();
}
$("import").addEventListener("click",()=>$("file").click());
$("file").addEventListener("change",e=>setMedia(e.target.files[0]));
for(const key of fields)$(key).addEventListener("input",e=>{project[key]=["scale","x","y"].includes(key)?Number(e.target.value):e.target.value;render()});
$("safe-toggle").addEventListener("change",e=>{project.safe=e.target.checked;render()});
$("reset").addEventListener("click",()=>{project.scale=80;project.x=0;project.y=0;render()});
$("play").addEventListener("click",()=>{if(!url)return;if(video.paused)video.play().catch(()=>notify("Video kan niet worden afgespeeld"));else video.pause()});
video.addEventListener("play",()=>$("play").textContent="Ⅱ");
video.addEventListener("pause",()=>$("play").textContent="▶");
video.addEventListener("loadedmetadata",()=>{render();updateClock()});
video.addEventListener("error",()=>notify("Videoformaat niet ondersteund in deze WebView"));
function updateClock(){
 const duration=Number.isFinite(video.duration)?video.duration:0;
 $("clock").textContent=timecode(video.currentTime)+" / "+timecode(duration);
 $("scrubber").value=duration?(video.currentTime/duration)*100:0;
}
video.addEventListener("timeupdate",updateClock);
$("scrubber").addEventListener("input",e=>{if(Number.isFinite(video.duration))video.currentTime=Number(e.target.value)/100*video.duration});
for(const [id,delta] of [["back",-5],["forward",5]])$(id).addEventListener("click",()=>{if(Number.isFinite(video.duration))video.currentTime=Math.max(0,Math.min(video.duration,video.currentTime+delta))});
$("save").addEventListener("click",()=>{
 const blob=new Blob([JSON.stringify(project,null,2)],{type:"application/json"});
 const a=document.createElement("a");a.href=URL.createObjectURL(blob);a.download="videolab-project.json";a.click();setTimeout(()=>URL.revokeObjectURL(a.href),1000);
 notify("Projectinstellingen opgeslagen; videobestand apart bewaren");
});
$("load").addEventListener("click",()=>$("project-file").click());
$("project-file").addEventListener("change",async e=>{
 const file=e.target.files[0];if(!file)return;
 try{const p=validate(JSON.parse(await file.text()));project=p;if(url){URL.revokeObjectURL(url);url=null};video.pause();video.removeAttribute("src");video.load();video.style.display="none";$("placeholder").hidden=false;$("scrubber").disabled=true;render();notify(p.media?"Project geopend — importeer "+p.media.name+" opnieuw":"Project geopend");}
 catch(err){notify(err.message)}
 e.target.value="";
});
render();
