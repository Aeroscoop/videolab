const google = "https://support.google.com/google-ads/answer/13547298#safe-zones";
const tiktok = "https://ads.tiktok.com/resources/help/article/tiktok-auction-in-feed-ads";
const linkedin = "https://www.linkedin.com/help/linkedin/answer/a7494039";

// Insets are top, right, bottom and left, expressed relative to the video canvas.
export const guides = {
 general: {label:"Algemeen", inset:[13,7,20,7]},
 "youtube-vertical": {
  label:"YouTube Ads · mobiel 9:16", format:"9:16", source:google,
  inset:[15,192/1080*100,35,48/1080*100],
 },
 "youtube-tablet": {
  label:"YouTube Ads · tablet 9:16", format:"9:16", source:google,
  inset:[14,18,31,9],
 },
 "youtube-wide": {
  label:"YouTube Ads · mobiel 16:9", format:"16:9", source:google,
  inset:[18,11,43,4],
 },
 "youtube-desktop": {
  label:"YouTube Ads · desktop 16:9", format:"16:9", source:google,
  inset:[17,9,44,4],
 },
 "tiktok-feed": {
  // Inscribed rectangle includes the wider right inset of the action column.
  label:"TikTok Ads · In-Feed LTR 9:16", format:"9:16", source:tiktok,
  inset:[240/1920*100,300/1080*100,660/1920*100,120/1080*100],
 },
 linkedin: {
  label:"LinkedIn · feed 4:5", format:"4:5", source:linkedin,
  inset:[8,8,12,8],
 },
};
