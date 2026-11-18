export const defaults = () => ({version:1,name:"Nieuw project",format:"4:5",background:"#a9c9b6",scale:80,x:0,y:0,safe:true,media:null});
export function validate(p){
 if(!p||p.version!==1||!["4:5","9:16","16:9","1:1"].includes(p.format))throw Error("Ongeldig VideoLab-project");
 const d=defaults();const out={...d,...p};
 for(const k of ["scale","x","y"]){if(typeof out[k]!=="number"||!Number.isFinite(out[k]))throw Error("Ongeldige waarde: "+k)}
 if(out.scale<20||out.scale>130||out.x< -100||out.x>100||out.y< -100||out.y>100)throw Error("Ongeldige schaal of positie");
 if(!/^#[0-9a-fA-F]{6}$/.test(out.background))throw Error("Ongeldige kleur");
 if(typeof out.safe!=="boolean")throw Error("Ongeldige richtzone");
 if(out.media!==null&&(typeof out.media!=="object"||typeof out.media.name!=="string"))throw Error("Ongeldige media");
 return out;
}
export function timecode(seconds){const s=Math.max(0,Math.floor(Number(seconds)||0));return String(Math.floor(s/60)).padStart(2,"0")+":"+String(s%60).padStart(2,"0")}
export function clipLabel(name){return (name||"Video").replace(/[<>]/g,"").slice(0,80)}
