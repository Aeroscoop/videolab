import {defaults,validate,timecode,clipLabel,minTrimDuration,validTrim,sourceClips,splitClips,nextClip,maxClips} from "./project.js";
import {guides} from "./guides.js";
const $=id=>document.getElementById(id);
let project=defaults(),url=null,busy=false;
let selectedClip=0,pendingField=null;
const undoStack=[];
const desktop=window.go?.main?.App;
let video=$("video");
const stage=$("stage"), videoLayer=$("video-layer");
const fields=["format","background","scale","x","y","speed","overlay"];
for(const [id,{label}] of Object.entries(guides)){
 $("overlay").add(new Option(label,id));
}
$("capture").hidden=!desktop;
$("capture").addEventListener("click",async ()=>{
 try{await desktop.OpenCapturePage()}
 catch(err){notify(err?.message??String(err))}
});
$("guide-source").addEventListener("click",()=>{
 const source=guides[project.overlay].source;
 if(desktop)window.runtime.BrowserOpenURL(source);
 else window.open(source,"_blank","noopener");
});
function remember(){
 undoStack.push(structuredClone(project));
 if(undoStack.length>100)undoStack.shift();
}
function undo(){
 if(!undoStack.length)return;
 project=undoStack.pop();
 selectedClip=Math.min(selectedClip,Math.max(0,project.clips.length-1));
 render();
 notify("Bewerking ongedaan gemaakt");
}
$("undo").addEventListener("click",undo);
window.addEventListener("keydown",event=>{
 if((event.ctrlKey||event.metaKey)&&!event.shiftKey&&event.key.toLowerCase()==="z"){
  event.preventDefault();
  undo();
 }
});
function notify(text){$("status").textContent=text}
function setBusy(value){
 busy=value;
 $("import").disabled=value;
 $("load").disabled=value;
 $("export").disabled=value||!desktop||!url;
}
function render(){
 $("export").disabled=busy||!desktop||!url;
 stage.className="stage ratio-"+project.format.replace(":","-");
 stage.style.background=project.background;
 $("background-label").textContent=project.background;
 const speed=$("speed");
 if(!Array.from(speed.options).some(option=>Number(option.value)===project.speed)){
  speed.add(new Option(project.speed+"×",project.speed));
 }
 for(const key of fields)$(key).value=project[key];
 for(const key of ["scale","x","y"])$(key+"-value").textContent=project[key]+(key==="scale"?"%":"%");
 $("safe-toggle").checked=project.safe;
 const guide=guides[project.overlay];
 const matching=!guide.format||guide.format===project.format;
 $("safe").style.inset=(matching?guide:guides.general).inset.map(value=>value+"%").join(" ");
 $("safe").classList.toggle("hidden",!project.safe);
 $("guide-source").hidden=!guide.source;
 video.style.transform="translate("+project.x+"%, "+project.y+"%) scale("+(project.scale/100)+")";
 $("clipname").textContent=project.media?.name||"Geen media";
 $("clip").textContent=project.media?clipLabel(project.media.name):"Video-track · importeer een opname";
 $("clip").classList.toggle("empty",!project.media);
 video.playbackRate=project.speed;
 const clips=sourceClips(project,video.duration);
 const length=clips.reduce((total,{start,end})=>total+end-start,0)/project.speed;
 $("duration").textContent=clips.length?trimTime(length)+" montage":"00:00";
 renderTrim(clips);
}
function resetMedia(){
 const previous=video;
 previous.pause();
 previous.removeAttribute("src");
 previous.load();
 video=previous.cloneNode(false);
 video.style.display="none";
 previous.replaceWith(video);
 if(url?.startsWith("blob:"))URL.revokeObjectURL(url);
 url=null;
 $("play").textContent="▶";
}
function setMedia({name,size,type},sourceURL){
 if(project.media?.name!==name||(Number.isFinite(project.media.size)&&project.media.size!==size)){
  project.trimStart=0;
  project.trimEnd=null;
  project.clips=[];
 }
 undoStack.length=0;
 selectedClip=0;
 if(url)resetMedia();
 url=sourceURL;
 video.src=url;
 video.style.display="block";
 $("placeholder").hidden=true;
 project.media={name,size,type};
 $("scrubber").disabled=false;
 notify("Media geïmporteerd");
 render();
}
$("import").addEventListener("click",async ()=>{
 if(!desktop){$("file").click();return}
 const previousStatus=$("status").textContent;
 setBusy(true);
 notify("Video openen…");
 try{
  const source=await desktop.OpenVideo();
  if(source)setMedia(source,source.url);
  else notify(previousStatus);
 }catch(err){notify(err?.message??String(err))}
 finally{setBusy(false)}
});
$("file").addEventListener("change",e=>{
 const input=e.currentTarget;
 const file=input.files[0];
 input.value="";
 if(!file||!file.type.startsWith("video/")){notify("Kies een videobestand");return}
 setMedia(file,URL.createObjectURL(file));
});
$("export").addEventListener("click",async ()=>{
 setBusy(true);
 notify("MP4 exporteren…");
 try{
  const name=await desktop.ExportMP4({sourceURL:url,project});
  notify(name?"MP4 opgeslagen: "+name:"Export geannuleerd");
 }catch(err){notify(err?.message??String(err))}
 finally{setBusy(false)}
});
for(const key of fields){
 $(key).addEventListener("input",event=>{
  if(pendingField!==key){remember();pendingField=key}
  project[key]=["scale","x","y","speed"].includes(key)?Number(event.target.value):event.target.value;
  const format=guides[project.overlay].format;
  if(key==="overlay"&&format){project.format=format;project.safe=true}
  if(key==="format"&&format&&format!==project.format)project.overlay="general";
  render();
 });
 $(key).addEventListener("change",()=>{pendingField=null});
}
$("safe-toggle").addEventListener("change",event=>{remember();project.safe=event.target.checked;render()});
$("reset").addEventListener("click",()=>{remember();project.scale=80;project.x=0;project.y=0;render()});
function trimTime(seconds){
 const milliseconds=Math.round(seconds*1000);
 const tenths=Math.floor(milliseconds/100)%10;
 return timecode(milliseconds/1000)+"."+tenths;
}
function renderTrim(clips){
 const ready=Boolean(url)&&Number.isFinite(video.duration)&&video.duration>=minTrimDuration&&clips.length>0;
 const edited=project.trimStart>0||project.trimEnd!==null||project.clips.length>0;
 $("trim-start").disabled=!ready;
 $("trim-end").disabled=!ready;
 $("split").disabled=!ready||clips.length>=maxClips;
 $("remove-clip").disabled=!ready||clips.length<2;
 $("trim-reset").disabled=!ready||!edited;
 $("undo").disabled=!undoStack.length;
 const clip=clips[selectedClip];
 $("trim-range").textContent=ready&&edited&&clip?trimTime(clip.start)+" – "+trimTime(clip.end):"";
 $("segments").hidden=!ready;
 $("clip").hidden=ready;
 if(ready)renderSegments(clips);
}
function renderSegments(clips){
 const buttons=clips.map(({start,end},i)=>{
  const button=document.createElement("button");
  button.className="segment";
  button.style.left=(start/video.duration*100)+"%";
  button.style.width=((end-start)/video.duration*100)+"%";
  button.textContent=clips.length===1?clipLabel(project.media.name):"Clip "+(i+1);
  button.title=trimTime(start)+" – "+trimTime(end);
  button.setAttribute("aria-pressed",String(i===selectedClip));
  button.addEventListener("click",()=>{
   selectedClip=i;
   video.currentTime=start;
   render();
  });
  return button;
 });
 $("segments").replaceChildren(...buttons);
}
function setTrim(start,end){
 const limit=end??video.duration;
 const previous=project.clips[selectedClip-1];
 const next=project.clips[selectedClip+1];
 if(!validTrim(start,limit)||limit>video.duration||(previous&&start<previous.end)||(next&&limit>next.start)){
  notify("Kies geldige grenzen binnen de bron, zonder andere clips te overlappen");
  return;
 }
 remember();
 if(project.clips.length){
  project.clips[selectedClip]={start,end:limit};
 }else{
  project.trimStart=start;
  project.trimEnd=end;
 }
 render();
 notify("Trimbereik ingesteld");
}
$("trim-start").addEventListener("click",()=>setTrim(video.currentTime,project.clips[selectedClip]?.end??project.trimEnd));
$("trim-end").addEventListener("click",()=>setTrim(project.clips[selectedClip]?.start??project.trimStart,video.currentTime));
$("trim-reset").addEventListener("click",()=>{
 remember();
 project.trimStart=0;
 project.trimEnd=null;
 project.clips=[];
 selectedClip=0;
 render();
 notify("Oorspronkelijk bereik hersteld");
});
$("split").addEventListener("click",()=>{
 const result=splitClips(sourceClips(project,video.duration),video.currentTime);
 if(!result){notify("Zet de afspeelkop binnen een clip om te splitsen");return}
 remember();
 project.clips=result.clips;
 project.trimStart=0;
 project.trimEnd=null;
 selectedClip=result.index;
 render();
 notify("Clip gesplitst");
});
$("remove-clip").addEventListener("click",()=>{
 if(project.clips.length<2)return;
 remember();
 project.clips.splice(selectedClip,1);
 selectedClip=Math.min(selectedClip,project.clips.length-1);
 video.currentTime=project.clips[selectedClip].start;
 render();
 notify("Clip uit montage verwijderd");
});
$("play").addEventListener("click",()=>{
 const clips=sourceClips(project,video.duration);
 if(!url||!clips.length)return;
 if(!video.paused){video.pause();return}
 const clip=nextClip(clips,video.currentTime)??clips[0];
 if(video.currentTime<clip.start||video.currentTime>=clip.end)video.currentTime=clip.start;
 video.play().catch(()=>notify("Video kan niet worden afgespeeld"));
});
function advancePlayback(){
 if(video.paused||video.seeking)return;
 const clips=sourceClips(project,video.duration);
 const clip=nextClip(clips,video.currentTime);
 if(clip){
  if(video.currentTime<clip.start)video.currentTime=clip.start;
 }else if(clips.length){
  video.pause();
  video.currentTime=clips.at(-1).end;
 }
}
function playbackFrame(){
 if(video.paused)return;
 advancePlayback();
 requestAnimationFrame(playbackFrame);
}
videoLayer.addEventListener("play",()=>{
 $("play").textContent="Ⅱ";
 requestAnimationFrame(playbackFrame);
},true);
videoLayer.addEventListener("pause",()=>$("play").textContent="▶",true);
videoLayer.addEventListener("loadedmetadata",()=>{
 const end=Math.min(project.trimEnd??video.duration,video.duration);
 if(project.clips.length){
  project.clips=project.clips
   .map(clip=>({...clip,end:Math.min(clip.end,video.duration)}))
   .filter(clip=>validTrim(clip.start,clip.end));
  selectedClip=0;
 }
 if((project.trimStart>0||project.trimEnd!==null)&&!validTrim(project.trimStart,end)){
  project.trimStart=0;
  project.trimEnd=null;
  notify("Trimbereik gewist: de video is korter dan het ingestelde bereik");
 }else if(project.trimEnd!==null){
  project.trimEnd=end;
 }
 const start=project.clips[0]?.start??project.trimStart;
 if(start>0)video.currentTime=start;
 render();
 updateClock();
},true);
videoLayer.addEventListener("error",()=>notify("Videoformaat niet ondersteund in deze WebView"),true);
function updateClock(){
 const duration=Number.isFinite(video.duration)?video.duration:0;
 const currentTime=duration?video.currentTime:0;
 $("clock").textContent=timecode(currentTime)+" / "+timecode(duration);
 $("scrubber").value=duration?(currentTime/duration)*100:0;
}
videoLayer.addEventListener("timeupdate",()=>{
 advancePlayback();
 updateClock();
},true);
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
 setBusy(true);
 try{
  const loaded=validate(JSON.parse(await file.text()));
  if(desktop)await desktop.ClearVideo();
  project=loaded;
  undoStack.length=0;
  selectedClip=0;
  if(url)resetMedia();
  $("placeholder").hidden=false;
  $("scrubber").disabled=true;
  render();
  updateClock();
  notify(project.media?"Project geopend — importeer "+project.media.name+" opnieuw":"Project geopend");
 }
 catch(err){notify(err?.message??String(err))}
 finally{e.target.value="";setBusy(false)}
});
render();
