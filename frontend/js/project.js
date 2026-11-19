import {guides} from "./guides.js";
export const defaults = () => ({version:1,name:"Nieuw project",format:"4:5",background:"#a9c9b6",scale:80,x:0,y:0,safe:true,media:null,trimStart:0,trimEnd:null,clips:[],speed:1,overlay:"general"});
export const minTrimDuration = 0.05;
export const maxClips = 64;
export function validTrim(start, end){
 return Number.isFinite(start)&&start>=0&&(end===null||(Number.isFinite(end)&&end-start>=minTrimDuration-0.000001));
}
export function validate(p){
 if(!p||p.version!==1||!["4:5","9:16","16:9","1:1"].includes(p.format))throw Error("Ongeldig VideoLab-project");
 const d=defaults();const out={...d,...p};
 for(const k of ["scale","x","y"]){if(typeof out[k]!=="number"||!Number.isFinite(out[k]))throw Error("Ongeldige waarde: "+k)}
 if(out.scale<20||out.scale>130||out.x< -100||out.x>100||out.y< -100||out.y>100)throw Error("Ongeldige schaal of positie");
 if(!/^#[0-9a-fA-F]{6}$/.test(out.background))throw Error("Ongeldige kleur");
 if(typeof out.safe!=="boolean")throw Error("Ongeldige richtzone");
 if(out.media!==null&&(typeof out.media!=="object"||typeof out.media.name!=="string"))throw Error("Ongeldige media");
 if(!Number.isFinite(out.speed)||out.speed<0.5||out.speed>2)throw Error("Ongeldige afspeelsnelheid");
 if(!Object.hasOwn(guides,out.overlay))throw Error("Ongeldige richtzone");
 if(!Array.isArray(out.clips)||out.clips.length>maxClips||out.clips.some((clip,i)=>
  !clip||!Number.isFinite(clip.end)||!validTrim(clip.start,clip.end)||(i>0&&clip.start<out.clips[i-1].end)
 ))throw Error("Ongeldige of overlappende clips");
 if(!validTrim(out.trimStart,out.trimEnd))throw Error("Ongeldige trimgrenzen");
 return out;
}
export function timecode(seconds){const s=Math.max(0,Math.floor(Number(seconds)||0));return String(Math.floor(s/60)).padStart(2,"0")+":"+String(s%60).padStart(2,"0")}
export function clipLabel(name){return (name||"Video").replace(/[<>]/g,"").slice(0,80)}

export function sourceClips(project,duration){
 if(project.clips.length)return project.clips;
 return Number.isFinite(duration)&&duration>0
  ? [{start:project.trimStart,end:project.trimEnd??duration}]
  : [];
}
export function splitClips(clips,at){
 const index=clips.findIndex(({start,end})=>validTrim(start,at)&&validTrim(at,end));
 if(index<0||clips.length>=maxClips)return null;
 const {start,end}=clips[index];
 return {index:index+1,clips:[...clips.slice(0,index),{start,end:at},{start:at,end},...clips.slice(index+1)]};
}
export function nextClip(clips,at){
 return clips.find(({end})=>at<end)??null;
}
