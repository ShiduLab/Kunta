'use strict';

const OPS = [
  ['Lettere indicate', 'Scrivi le lettere da contare, es. aeiou'],
  ['Caratteri', 'Conta caratteri Unicode e mostra la frequenza di ogni carattere'],
  ['Parole', 'Conta parole, parole distinte e mostra le frequenze'],
  ['Righe', 'Conta righe e righe non vuote'],
  ['Spazi', 'Conta spazi semplici e caratteri bianchi'],
  ['Cifre', 'Conta le cifre'],
  ['Punteggiatura', 'Conta i segni di punteggiatura'],
  ['Occorrenze parola/stringa', 'Parametro: parola o frase da cercare'],
  ['Parole uniche', 'Elenca le parole distinte'],
  ['Parole ripetute', 'Elenca solo le parole che ricorrono'],
  ['Frequenza lettere', 'Classifica le lettere per frequenza'],
  ['Frequenza parole', 'Classifica le parole per frequenza'],
  ['Parola più corta / più lunga', 'Trova gli estremi di lunghezza'],
  ['Palindromo puro', 'Parametro: lunghezza minima, default 3'],
  ['Bifronti', 'Parametro: lunghezza minima, default 3'],
  ['Anagrammi', 'Cerca nel dizionario italiano; parametro: lunghezza minima, default 3'],
  ['Acrostico', 'Prima lettera/cifra utile di ogni riga non vuota'],
  ['Telestico', 'Ultima lettera/cifra utile di ogni riga non vuota'],
  ['Inverti testo', 'Inverte tutti i caratteri'],
  ['Inverti ordine parole', 'Inverte la sequenza delle parole'],
  ['Ordina parole alfabeticamente', 'Visualizza subito ripetizioni e famiglie lessicali'],
  ['Estrai numeri', 'Estrae le sequenze numeriche'],
  ['Schema rime / desinenze', 'Parametro: profondità della coda, default 3'],
  ['Rima baciata / inclusione progressiva', 'Parametro: nucleo minimo, default 3'],
  ['LetterTransport', 'Parametro: trasporto minimo, default 3'],
  ['Sciarade / salti di dominio', 'Parametro: lunghezza minima segmento, default 2'],
  ['Kunta il testo (fenomeni)', 'Ricognizione combinata di strutture e ricorrenze'],
  ['Isovocaliche', 'Raggruppa parole con lo stesso scheletro vocalico'],
  ['Isoconsonantiche', 'Raggruppa parole con lo stesso scheletro consonantico'],
  ['Omovocaliche', 'Stesso materiale vocalico; mostra le vocali nell’ordine reale di ogni parola'],
  ['Omoconsonantiche', 'Stesso materiale consonantico; mostra le consonanti nell’ordine reale di ogni parola'],
  ['Omovocaliche iniziali', 'Parametro: quante vocali iniziali confrontare, default 1'],
  ['Omovocaliche finali', 'Parametro: quante vocali finali confrontare, default 1'],
  ['Omoconsonantiche iniziali', 'Parametro: quante consonanti iniziali confrontare, default 1'],
  ['Omoconsonantiche finali', 'Parametro: quante consonanti finali confrontare, default 1'],
  ['Lunghezza media parole', 'Calcola la lunghezza media delle parole'],
  ['Distribuzione lunghezze', 'Conta quante parole hanno 1, 2, 3… lettere'],
  ['Parole con iniziale', 'Parametro: una o più lettere iniziali'],
  ['Parole con finale', 'Parametro: una o più lettere finali'],
  ['Parole contenenti sequenza', 'Parametro: sequenza da cercare dentro le parole'],
  ['Parole di lunghezza N', 'Parametro: numero esatto di lettere'],
  ['Doppie / triple', 'Rileva lettere consecutive ripetute nelle parole'],
  ['Sequenze ripetute', 'Parametro: lunghezza minima della sequenza, default 2'],
  ['Isogrammi', 'Parole senza lettere ripetute'],
  ['Parole alfabetiche', 'Lettere consecutive dell’alfabeto: AB, ABC, BCD…'],
  ['Parole alfabetiche inverse', 'Lettere consecutive dell’alfabeto al contrario: BA, CBA, FED…'],
  ['Elimina duplicati', 'Restituisce le parole una sola volta, nell’ordine di apparizione'],
  ['Estrai parole', 'Estrae solo le parole dal testo'],
  ['Testo in MAIUSCOLO', 'Converte il testo in maiuscolo'],
  ['Testo in minuscolo', 'Converte il testo in minuscolo'],
  ['Palindromo inverso', 'Parola letta al contrario; mostra solo coppie realmente presenti nel testo'],
  ['Palindromo contrario', 'Rileva forme tipo POSSESSO: prima lettera fissa, resto palindromo'],
  ['Inversi', 'Ultima lettera fissa; mostra solo coppie realmente presenti nel testo'],
  ['Antipodi', 'Prima lettera fissa; mostra solo coppie realmente presenti nel testo'],
  ['Allitterazioni', 'Raggruppa parole per lettera iniziale ricorrente'],
  ['Assonanze', 'Raggruppa parole per coda vocalica'],
  ['Consonanze', 'Raggruppa parole per coda consonantica'],
  ['Ossimori — candidati', 'Cerca coppie di termini semanticamente contrari o paradossali vicini']
];

const $ = id => document.getElementById(id);
const input = $('inputText'), output = $('outputText'), opSel = $('operation'), param = $('param');

OPS.forEach(([name], i) => {
  const o = document.createElement('option'); o.value = i; o.textContent = name; opSel.appendChild(o);
});
opSel.value = '20';

function updateHint(){ $('hint').textContent = OPS[Number(opSel.value)][1]; }
opSel.addEventListener('change', updateHint); updateHint();

function normalizeCase(s, sensitive){ return sensitive ? s : s.toLowerCase(); }
function normalizeLines(s){ return s.replace(/\r\n?/g,'\n').split('\n'); }
function tokenizeWords(s){ return s.match(/[\p{L}\p{N}]+(?:[’'\-][\p{L}\p{N}]+)*/gu) || []; }
function cleanWord(s, sensitive){
  let w = s.replace(/^[^\p{L}\p{N}]+|[^\p{L}\p{N}]+$/gu,'');
  return normalizeCase(w, sensitive);
}
function reverseRunes(s){ return Array.from(s).reverse().join(''); }
function parsePositive(s, def, min, max){ const n = Number.parseInt(String(s).trim(),10); return Number.isFinite(n) ? Math.max(min,Math.min(max,n)) : def; }
function parseMinLen(s){ return parsePositive(s,3,1,100); }
function limitLines(s,max){ return s; }
function freqWords(words, sensitive){
  const freq=new Map(), display=new Map();
  for(const w0 of words){ const k=cleanWord(w0,sensitive); if(!k) continue; freq.set(k,(freq.get(k)||0)+1); if(!display.has(k)) display.set(k,w0); }
  return {freq,display};
}
function distinctWordCount(words,sensitive){ return freqWords(words,sensitive).freq.size; }
function sortedEntries(map, cmp){ return [...map.entries()].sort(cmp); }
function countOccurrences(hay,needle){ if(!needle) return 0; let n=0,pos=0; while((pos=hay.indexOf(needle,pos))!==-1){n++;pos+=needle.length;} return n; }

function findPalindromes(words,sensitive,minLen){
  const {freq,display}=freqWords(words,sensitive); const rows=[]; let total=0;
  for(const [k,n] of freq){ if(Array.from(k).length>=minLen && k===reverseRunes(k)){rows.push([k,n]);total+=n;} }
  rows.sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0]));
  return `Palindromi distinti: ${rows.length}\nOccorrenze totali: ${total}\nLunghezza minima: ${minLen}\n\n`+rows.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n');
}
function findBifronti(words,sensitive,minLen){
  const {freq,display}=freqWords(words,sensitive); const seen=new Set(), pairs=[];
  for(const k of freq.keys()){
    if(Array.from(k).length<minLen) continue; const rev=reverseRunes(k); if(rev===k||!freq.has(rev)) continue;
    const key=[k,rev].sort().join('\0'); if(seen.has(key)) continue; seen.add(key); pairs.push([display.get(k),display.get(rev)]);
  }
  pairs.sort((a,b)=>a[0].localeCompare(b[0],undefined,{sensitivity:'base'}));
  return `Bifronti trovati: ${pairs.length}\nLunghezza minima: ${minLen}\n\n`+pairs.map(p=>`${p[0]} ↔ ${p[1]}`).join('\n');
}
function sortedRunes(s){ return Array.from(s).sort().join(''); }
function baseAnagramChar(ch){
  const c=ch.toLowerCase();
  if('aàáâäãå'.includes(c)) return 'a';
  if('eèéêë'.includes(c)) return 'e';
  if('iìíîï'.includes(c)) return 'i';
  if('oòóôöõ'.includes(c)) return 'o';
  if('uùúûü'.includes(c)) return 'u';
  if(c==='ç') return 'c';
  return c>='a'&&c<='z'?c:'';
}
function anagramIdentity(word){
  let out='';
  for(const ch of Array.from(word)){
    if(/\p{N}/u.test(ch)) return '';
    out+=baseAnagramChar(ch);
  }
  return out;
}
function anagramSignatureKey(word){
  const counts=new Uint8Array(26); let letters=0;
  for(const ch of Array.from(word)){
    if(/\p{N}/u.test(ch)) return ['',0];
    const b=baseAnagramChar(ch); if(!b) continue;
    counts[b.charCodeAt(0)-97]++; letters++;
  }
  if(!letters) return ['',0];
  return [[...counts].map(n=>n.toString(16).padStart(2,'0')).join(''),letters];
}
function anagramLexeme(raw){
  let w=cleanWord(raw,true);
  const i=Math.max(w.lastIndexOf("'"),w.lastIndexOf('’'));
  if(i>=0 && i+1<w.length) w=w.slice(i+1);
  return cleanWord(w,true);
}
const anagramResultCache=new Map();
async function dictionaryMatches(signatures){
  const need=[...signatures].filter(sig=>!anagramResultCache.has(sig));
  if(!need.length){ const out=new Map(); for(const sig of signatures) out.set(sig,anagramResultCache.get(sig)||[]); return out; }
  const wanted=new Set(need);
  let found={};
  if(window.KuntaNative && typeof window.KuntaNative.lookupAnagrams==='function'){
    const raw=String(window.KuntaNative.lookupAnagrams(JSON.stringify(need))||'{}');
    found=JSON.parse(raw);
    if(found.__error) throw new Error(found.__error);
  }else{
    if(typeof DecompressionStream==='undefined' || typeof TextDecoderStream==='undefined') throw new Error('Il browser non supporta la decompressione del dizionario.');
    const resp=await fetch('./assets/anagram_index.tsv.gz');
    if(!resp.ok) throw new Error(`Dizionario non disponibile (${resp.status})`);
    const stream=resp.body.pipeThrough(new DecompressionStream('gzip')).pipeThrough(new TextDecoderStream());
    const reader=stream.getReader(); let carry='';
    while(true){
      const {value,done}=await reader.read();
      if(value) carry+=value;
      let pos;
      while((pos=carry.indexOf('\n'))>=0){
        const line=carry.slice(0,pos); carry=carry.slice(pos+1);
        const tab=line.indexOf('\t'); if(tab<=0) continue;
        const sig=line.slice(0,tab); if(wanted.has(sig)) found[sig]=line.slice(tab+1).split('|');
      }
      if(done) break;
    }
    if(carry){ const tab=carry.indexOf('\t'); if(tab>0){const sig=carry.slice(0,tab);if(wanted.has(sig))found[sig]=carry.slice(tab+1).split('|');} }
  }
  for(const sig of need) anagramResultCache.set(sig,Array.isArray(found[sig])?found[sig]:[]);
  const out=new Map(); for(const sig of signatures) out.set(sig,anagramResultCache.get(sig)||[]); return out;
}
async function findAnagrams(words,sensitive,minLen){
  const targets=new Set(), wordSig=new Map(), display=new Map(), ordered=[];
  for(const raw of words){
    const lexeme=anagramLexeme(raw); if(!lexeme) continue;
    const k=normalizeCase(lexeme,sensitive); if(display.has(k)) continue;
    const [sig,letters]=anagramSignatureKey(k); if(!sig||letters<minLen) continue;
    targets.add(sig); wordSig.set(k,sig); display.set(k,lexeme); ordered.push(k);
  }
  if(!targets.size) return `Anagrammi trovati: 0\nLunghezza minima: ${minLen}\n\nNessuna parola analizzabile.`;
  const matches=await dictionaryMatches(targets);
  const out=[]; let groups=0;
  for(const k of ordered){
    const candidates=matches.get(wordSig.get(k))||[]; if(!candidates.length) continue;
    const sourceNorm=anagramIdentity(display.get(k)), seen=new Set(), filtered=[];
    for(const candidate of candidates){
      const cn=anagramIdentity(candidate); if(!cn||cn===sourceNorm||seen.has(cn)) continue;
      seen.add(cn); filtered.push(candidate);
    }
    if(!filtered.length) continue;
    filtered.sort((a,b)=>a.localeCompare(b,'it',{sensitivity:'base'}));
    if(groups) out.push('');
    out.push(`${display.get(k)} →`); filtered.forEach(x=>out.push(`    ${x}`)); groups++;
  }
  const head=[`Anagrammi: ${groups} parole del testo con almeno un anagramma`,`Lunghezza minima: ${minLen}`,''];
  if(!groups) head.push('Nessun anagramma trovato nel dizionario.');
  return [...head,...out].join('\n');
}
function runeSuffix(s,n){ const r=Array.from(s); return r.length<=n?s:r.slice(-n).join(''); }
function finalWordOfLine(line,sensitive){ const w=tokenizeWords(line); return w.length?cleanWord(w[w.length-1],sensitive):''; }
function rhymeLabel(n){ const labels='xyzwvutsrqponmlkjihgfedcba'; return n<labels.length?labels[n]:`r${n+1}`; }
function analyzeRhymeScheme(text,sensitive,depth){
  const lines=normalizeLines(text), stanzas=[]; let cur=[];
  const flush=()=>{if(cur.length){stanzas.push(cur);cur=[];}};
  for(const line of lines){ if(!line.trim()) flush(); else cur.push(line); } flush();
  if(!stanzas.length) return 'Nessun verso rilevato.';
  const out=[`Rima = desinenza grafica · profondità: ${depth} lettere`,`Strofe rilevate: ${stanzas.length}`,''];
  stanzas.forEach((stanza,si)=>{
    const rows=[], counts=new Map(), order=[];
    for(const line of stanza){ const word=finalWordOfLine(line,sensitive); if(!word) continue; const suffix=runeSuffix(word,depth); rows.push({word,suffix}); counts.set(suffix,(counts.get(suffix)||0)+1); if(!order.includes(suffix)) order.push(suffix); }
    if(!rows.length) return;
    const groups=new Map();
    if(order.length===2){ let singleton='',dominant=''; for(const s of order){ if(counts.get(s)===1) singleton=s; else dominant=s; } if(singleton&&dominant){groups.set(dominant,'x');groups.set(singleton,'z');} }
    if(!groups.size) order.forEach((s,i)=>groups.set(s,rhymeLabel(i)));
    rows.forEach(r=>r.label=groups.get(r.suffix));
    if(stanzas.length>1) out.push(`STROFA ${si+1}`);
    rows.forEach((r,i)=>out.push(`${String(i+1).padStart(2)}. ${r.word.padEnd(22)} → -${r.suffix.padEnd(10)} → ${r.label}`));
    out.push('',`Schema: ${rows.map(r=>r.label).join('-')}`,'Coda progressiva:');
    for(let d=1;d<=depth;d++){
      const f=new Map(); rows.forEach(r=>{const s=runeSuffix(r.word,d);f.set(s,(f.get(s)||0)+1)});
      const parts=[...f.entries()].filter(([,n])=>n>1).sort((a,b)=>a[0].localeCompare(b[0])).map(([s,n])=>`-${s} ×${n}`);
      if(parts.length) out.push(`  ${d}: ${parts.join(' · ')}`);
    }
    if(si<stanzas.length-1) out.push('');
  });
  return limitLines(out.join('\n'),700);
}
function findInclusionChains(words,sensitive,minLen){
  const {display}=freqWords(words,sensitive); const keys=[...display.keys()].filter(k=>Array.from(k).length>=minLen).sort((a,b)=>Array.from(a).length-Array.from(b).length||a.localeCompare(b));
  const children=new Map(), incoming=new Map();
  for(const a of keys){
    const la=Array.from(a).length; const candidates=keys.filter(b=>Array.from(b).length>la&&b.endsWith(a));
    for(const b of candidates){ const lb=Array.from(b).length; let immediate=true;
      for(const c of candidates){ const lc=Array.from(c).length; if(c!==b&&lc>la&&lc<lb&&c.endsWith(a)&&b.endsWith(c)){immediate=false;break;} }
      if(immediate){ if(!children.has(a)) children.set(a,[]); children.get(a).push(b); incoming.set(b,(incoming.get(b)||0)+1); }
    }
    if(children.has(a)) children.get(a).sort((x,y)=>Array.from(x).length-Array.from(y).length||x.localeCompare(y));
  }
  const chains=[];
  function dfs(node,path){ const p=[...path,node], ch=children.get(node)||[]; if(!ch.length){if(p.length>=2) chains.push(p);return;} ch.forEach(c=>dfs(c,p)); }
  keys.filter(k=>!(incoming.get(k)>0)&&(children.get(k)||[]).length).forEach(k=>dfs(k,[]));
  chains.sort((a,b)=>b.length-a.length||a[0].localeCompare(b[0]));
  const out=[`Rima baciata · inclusione progressiva`,`Nucleo minimo: ${minLen} lettere`,`Catene trovate: ${chains.length}`,''];
  chains.slice(0,100).forEach((c,i)=>{out.push(`${String(i+1).padStart(2)}. ${c.map(k=>display.get(k)).join(' → ')}`,`    nucleo: ${display.get(c[0])} · profondità: ${c.length}`)});
  if(!chains.length) out.push('Nessuna inclusione progressiva rilevata.');
  if(chains.length>100) out.push('\n… altre catene omesse.');
  return out.join('\n');
}
function longestCommonSubstring(a,b){
  const ar=Array.from(a), br=Array.from(b); if(!ar.length||!br.length) return '';
  let prev=new Array(br.length+1).fill(0), bestLen=0,bestEnd=0;
  for(let i=1;i<=ar.length;i++){ const cur=new Array(br.length+1).fill(0); for(let j=1;j<=br.length;j++){ if(ar[i-1]===br[j-1]){cur[j]=prev[j-1]+1;if(cur[j]>bestLen){bestLen=cur[j];bestEnd=i;}} } prev=cur; }
  return ar.slice(bestEnd-bestLen,bestEnd).join('');
}
function findLetterTransport(words,sensitive,minLen){
  if(words.length<2) return 'Servono almeno due parole.'; const out=[`LetterTransport · parole consecutive`,`Trasporto minimo: ${minLen} lettere`,'']; let found=0;
  for(let i=0;i+1<words.length&&found<200;i++){ const a0=words[i],b0=words[i+1],a=cleanWord(a0,sensitive),b=cleanWord(b0,sensitive); if(!a||!b||a===b) continue; const common=longestCommonSubstring(a,b); const n=Array.from(common).length; if(n<minLen) continue; found++; out.push(`${String(found).padStart(3)}. ${a0} → ${b0}`,`     trasporta: "${common}" (${n} lettere)`); }
  if(!found) out.push('Nessun trasporto consecutivo rilevato.'); if(found>=200) out.push('\n… output limitato a 200 passaggi.'); return out.join('\n');
}
function phraseExists(tokens,parts){ outer:for(let i=0;i+parts.length<=tokens.length;i++){for(let j=0;j<parts.length;j++) if(tokens[i+j]!==parts[j]) continue outer; return true;} return false; }
function findSciarades(text,sensitive,minSeg){
  const raw=tokenizeWords(text), tokens=[], counts=new Map(), display=new Map();
  for(const w of raw){const k=cleanWord(w,sensitive);if(!k)continue;tokens.push(k);counts.set(k,(counts.get(k)||0)+1);if(!display.has(k))display.set(k,w);}
  const targets=[...counts.keys()].filter(k=>Array.from(k).length>=minSeg*2).sort((a,b)=>Array.from(b).length-Array.from(a).length||a.localeCompare(b));
  const out=[`Sciarade interne · segmenti presenti nel testo`,`Segmento minimo: ${minSeg} lettere`,'']; let found=0;
  for(const target of targets){ const rr=Array.from(target), segSet=new Set(), segs=[];
    for(let i=minSeg;i<=rr.length-minSeg;i++){ const a=rr.slice(0,i).join(''),b=rr.slice(i).join(''); if(!counts.has(a)||!counts.has(b)||!phraseExists(tokens,[a,b])) continue; const key=`${a}|${b}`;if(!segSet.has(key)){segSet.add(key);segs.push(key);} }
    for(let i=minSeg;i<=rr.length-2*minSeg;i++) for(let j=i+minSeg;j<=rr.length-minSeg;j++){ const a=rr.slice(0,i).join(''),b=rr.slice(i,j).join(''),c=rr.slice(j).join(''); if(!counts.has(a)||!counts.has(b)||!counts.has(c)||!phraseExists(tokens,[a,b,c]))continue; const key=`${a}|${b}|${c}`;if(!segSet.has(key)){segSet.add(key);segs.push(key);} }
    if(!segs.length) continue; found++; let head=display.get(target); if((counts.get(target)||0)>1) head+=`  [forma intera ×${counts.get(target)}]`; out.push(head);
    segs.sort().forEach(seg=>out.push('  → '+seg.split('|').map(p=>display.get(p)||p).join(' + ')));
    if((counts.get(target)||0)>1) out.push('  ↳ possibile salto di dominio: stessa forma intera + segmentazioni diverse; verifica il senso nel contesto.'); out.push('');
    if(found>=100){out.push('… altri candidati omessi.');break;}
  }
  if(!found) out.push('Nessuna sciarada interna rilevata con il vocabolario del testo.'); return out.join('\n');
}
function repeatedWordsSummary(text,sensitive,maxItems){
  const {freq,display}=freqWords(tokenizeWords(text),sensitive); const a=[...freq.entries()].filter(([,n])=>n>1).sort((x,y)=>y[1]-x[1]||x[0].localeCompare(y[0])).slice(0,maxItems); return a.length?a.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'):'Nessuna parola ripetuta.';
}
function kuntaPhenomena(text,sensitive){ return limitLines([
  'KUNTA IL TESTO','==============================','','RIPETIZIONI',repeatedWordsSummary(text,sensitive,20),'','SCHEMA RIME / DESINENZE',analyzeRhymeScheme(text,sensitive,3),'','RIMA BACIATA / INCLUSIONE PROGRESSIVA',findInclusionChains(tokenizeWords(text),sensitive,3),'','LETTERTRANSPORT',findLetterTransport(tokenizeWords(text),sensitive,3),'','SCIARADE / SALTI DI DOMINIO',findSciarades(text,sensitive,2),'','ISOVOCALICHE',groupWordSignatures(tokenizeWords(text),'Scheletro vocalico',vowelSkeleton),'','ISOCONSONANTICHE',groupWordSignatures(tokenizeWords(text),'Scheletro consonantico',consonantSkeleton)
].join('\n'),1100); }


function foldItalianVowel(ch){
  const c=ch.toLowerCase();
  if('aàáâäãå'.includes(c)) return 'a';
  if('eèéêë'.includes(c)) return 'e';
  if('iìíîï'.includes(c)) return 'i';
  if('oòóôöõ'.includes(c)) return 'o';
  if('uùúûü'.includes(c)) return 'u';
  return '';
}
function vowelSkeleton(word){
  let out='';
  for(const ch of Array.from(cleanWord(word,false))){ const v=foldItalianVowel(ch); if(v) out+=v; }
  return out;
}
function consonantSkeleton(word){
  let out='';
  for(const ch of Array.from(cleanWord(word,false))){ if(/\p{L}/u.test(ch) && !foldItalianVowel(ch)) out+=ch.toLowerCase(); }
  return out;
}
function sortedSignature(sig){ return Array.from(sig).sort().join(''); }
function edgeSignature(sig,n,fromEnd){ const r=Array.from(sig); n=Math.max(1,n|0); if(r.length<n)return ''; return (fromEnd?r.slice(-n):r.slice(0,n)).join(''); }
function prettySignature(sig){ return Array.from(sig.toUpperCase()).join('-'); }
function groupWordSignatures(words,label,signature){
  const groups=new Map(), seen=new Map();
  for(const raw of words){
    const clean=cleanWord(raw,false); if(!clean) continue;
    const key=signature(raw); if(!key) continue;
    if(!groups.has(key)){groups.set(key,[]);seen.set(key,new Set());}
    if(!seen.get(key).has(clean)){seen.get(key).add(clean);groups.get(key).push(raw);}
  }
  const keys=[...groups.keys()].filter(k=>groups.get(k).length>=2).sort((a,b)=>Array.from(b).length-Array.from(a).length||a.localeCompare(b));
  const out=[`${label}: ${keys.length} gruppi`,''];
  for(const k of keys){ out.push(`${prettySignature(k)}  →`); for(const item of groups.get(k)) out.push(`    ${item}`); out.push(''); }
  if(!keys.length) out.push('Nessun gruppo rilevato nel testo.');
  return limitLines(out.join('\n'),700);
}
function materialCountLabel(s){
  const m=new Map(); for(const ch of Array.from(s.toUpperCase())) m.set(ch,(m.get(ch)||0)+1);
  return [...m.entries()].sort((a,b)=>a[0].localeCompare(b[0])).map(([ch,n])=>`${ch}×${n}`).join(' ');
}
function groupWordMaterials(words,label,orderedSig){
  const groups=new Map(),seen=new Map();
  for(const raw of words){ const clean=cleanWord(raw,false);if(!clean)continue;const ordered=orderedSig(raw);if(!ordered)continue;const material=sortedSignature(ordered);if(!groups.has(material)){groups.set(material,[]);seen.set(material,new Set());}if(!seen.get(material).has(clean)){seen.get(material).add(clean);groups.get(material).push(raw);} }
  const keys=[...groups.keys()].filter(k=>groups.get(k).length>=2).sort((a,b)=>Array.from(b).length-Array.from(a).length||a.localeCompare(b));
  const out=[`${label}: ${keys.length} gruppi`,''];
  for(const k of keys){ out.push(`Materiale: ${materialCountLabel(k)}`); for(const item of groups.get(k)) out.push(`    ${item}  →  ${prettySignature(orderedSig(item))}`); out.push(''); }
  if(!keys.length) out.push('Nessun gruppo rilevato nel testo.');
  return limitLines(out.join('\n'),700);
}

function wordLen(w){ return Array.from(cleanWord(w,true)).length; }
function uniqueDisplayWords(words,sensitive){
  const {display}=freqWords(words,sensitive); return [...display.values()];
}
function filterWordsBy(words,p,sensitive,mode){
  const q=normalizeCase(p.trim(),sensitive); if(!q)return 'Scrivi il parametro da cercare.';
  const a=uniqueDisplayWords(words,sensitive).filter(w=>{
    const k=normalizeCase(cleanWord(w,true),sensitive);
    return mode==='start'?k.startsWith(q):mode==='end'?k.endsWith(q):k.includes(q);
  });
  return `${a.length} parole trovate\n\n`+a.join('\n');
}
function findRepeatedRuns(words){
  const out=[];
  for(const w of uniqueDisplayWords(words,false)){
    const rr=Array.from(cleanWord(w,false)); const runs=[];
    for(let i=0;i<rr.length;){ let j=i+1; while(j<rr.length&&rr[j]===rr[i])j++; if(j-i>=2)runs.push(rr.slice(i,j).join('')); i=j; }
    if(runs.length) out.push(`${w}  →  ${runs.join(' · ')}`);
  }
  return `Parole con doppie/triple: ${out.length}\n\n`+out.join('\n');
}
function findRepeatedSequences(words,minLen){
  const out=[];
  for(const w of uniqueDisplayWords(words,false)){
    const a=Array.from(cleanWord(w,false)); const found=new Set();
    for(let n=minLen;n<=Math.floor(a.length/2);n++){
      for(let i=0;i+n<=a.length;i++){
        const seq=a.slice(i,i+n).join(''); let c=0;
        for(let j=0;j+n<=a.length;j++) if(a.slice(j,j+n).join('')===seq)c++;
        if(c>1)found.add(seq);
      }
    }
    if(found.size) out.push(`${w}  →  ${[...found].sort((x,y)=>y.length-x.length||x.localeCompare(y)).join(' · ')}`);
  }
  return `Parole con sequenze ripetute: ${out.length}\nLunghezza minima: ${minLen}\n\n`+out.join('\n');
}
function isIsogram(w){ const a=Array.from(cleanWord(w,false)).filter(ch=>/\p{L}/u.test(ch)); return a.length>1&&new Set(a).size===a.length; }
function foldAlphabetChar(ch){ const c=ch.toLowerCase(); if('aàáâäãå'.includes(c))return'a';if('eèéêë'.includes(c))return'e';if('iìíîï'.includes(c))return'i';if('oòóôöõ'.includes(c))return'o';if('uùúûü'.includes(c))return'u';return c; }
function alphabeticWord(w,descending=false){
  const a=Array.from(cleanWord(w,false)).filter(ch=>/\p{L}/u.test(ch)).map(foldAlphabetChar);
  if(a.length<2||a.some(ch=>ch<'a'||ch>'z'))return false;
  const step=descending?-1:1;
  for(let i=1;i<a.length;i++) if(a[i].charCodeAt(0)!==a[i-1].charCodeAt(0)+step)return false;
  return true;
}
function findReversePairs(words,sensitive,minLen,label){
  const {freq,display}=freqWords(words,sensitive),seen=new Set(),pairs=[];
  for(const k of freq.keys()){ if(Array.from(k).length<minLen)continue;const rev=reverseRunes(k);if(rev===k||!freq.has(rev))continue;const key=[k,rev].sort().join('\0');if(seen.has(key))continue;seen.add(key);pairs.push([display.get(k),display.get(rev)]); }
  pairs.sort((a,b)=>a[0].localeCompare(b[0],'it',{sensitivity:'base'}));
  const out=[`${label} trovati: ${pairs.length}`,`Lunghezza minima: ${minLen}`,'']; pairs.forEach(([a,b])=>out.push(`${a} ↔ ${b}`));if(!pairs.length)out.push('Nessuna coppia presente nel testo.');return out.join('\n');
}
function transformInverse(w){ const r=Array.from(cleanWord(w,true));if(r.length<2)return'';return r.slice(0,-1).reverse().join('')+r.at(-1); }
function transformAntipode(w){ const r=Array.from(cleanWord(w,true));if(r.length<2)return'';return r[0]+r.slice(1).reverse().join(''); }
function findTransformPairs(words,sensitive,minLen,label,transform){
  const {freq,display}=freqWords(words,sensitive),seen=new Set(),pairs=[];
  for(const k of freq.keys()){ if(Array.from(k).length<minLen)continue;const other=normalizeCase(transform(k),sensitive);if(!other||other===k||!freq.has(other))continue;const key=[k,other].sort().join('\0');if(seen.has(key))continue;seen.add(key);pairs.push([display.get(k),display.get(other)]); }
  pairs.sort((a,b)=>a[0].localeCompare(b[0],'it',{sensitivity:'base'}));
  const out=[`${label} trovati: ${pairs.length}`,`Lunghezza minima: ${minLen}`,''];pairs.forEach(([a,b])=>out.push(`${a} ↔ ${b}`));if(!pairs.length)out.push('Nessuna coppia presente nel testo.');return out.join('\n');
}
function displayCharacter(ch,sensitive){
  let c=ch; if(!sensitive&&/\p{L}/u.test(c))c=c.toLocaleUpperCase('it');
  if(c===' ')return'[spazio]';if(c==='\n')return'[a capo]';if(c==='\r')return'[ritorno carrello]';if(c==='\t')return'[tab]';return c;
}
function characterFrequency(text,sensitive){
  const f=new Map();for(const ch of Array.from(text)){const k=!sensitive&&/\p{L}/u.test(ch)?ch.toLocaleLowerCase('it'):ch;f.set(k,(f.get(k)||0)+1);}
  const keys=[...f.keys()].sort((a,b)=>f.get(b)-f.get(a)||displayCharacter(a,sensitive).localeCompare(displayCharacter(b,sensitive),'it'));
  return [`Caratteri: ${Array.from(text).length}`,'',...keys.map(k=>`${f.get(k)}: ${displayCharacter(k,sensitive)}`)].join('\n');
}
function wordFrequencySummary(words,sensitive){
  const {freq,display}=freqWords(words,sensitive);const keys=[...freq.keys()].sort((a,b)=>freq.get(b)-freq.get(a)||a.localeCompare(b,'it'));
  return [`Parole: ${words.length}`,`Parole distinte: ${keys.length}`,'',...keys.map(k=>`${freq.get(k)}: ${display.get(k)}`)].join('\n');
}
function findContraryPalindromes(words,minLen){
  const out=[];
  for(const w of uniqueDisplayWords(words,false)){
    const k=cleanWord(w,false), r=Array.from(k); if(r.length<minLen+1)continue;
    const tail=r.slice(1).join(''); if(tail===reverseRunes(tail))out.push(w);
  }
  return `Palindromi contrari: ${out.length}\n\n`+out.join('\n');
}
function groupByKey(words,label,keyFn,minGroup=2){
  const g=new Map(),seen=new Map();
  for(const w of words){ const clean=cleanWord(w,false); if(!clean)continue; const k=keyFn(w); if(!k)continue; if(!g.has(k)){g.set(k,[]);seen.set(k,new Set());} if(!seen.get(k).has(clean)){seen.get(k).add(clean);g.get(k).push(w);} }
  const rows=[...g.entries()].filter(([,a])=>a.length>=minGroup).sort((a,b)=>b[1].length-a[1].length||a[0].localeCompare(b[0]));
  return `${label}: ${rows.length} gruppi\n\n`+rows.map(([k,a])=>`${k.toUpperCase()}  →  ${a.join(' · ')}`).join('\n');
}
function suffixChars(sig,n){ const a=Array.from(sig); return a.slice(-Math.min(n,a.length)).join(''); }
function findOxymoronCandidates(text){
  const pairs=[
    ['vivo','morto'],['vita','morte'],['caldo','freddo'],['luce','buio'],['chiaro','scuro'],['vero','falso'],['pieno','vuoto'],['grande','piccolo'],['alto','basso'],['forte','debole'],['ricco','povero'],['aperto','chiuso'],['vicino','lontano'],['vecchio','nuovo'],['pace','guerra'],['amore','odio'],['ordine','caos'],['presenza','assenza'],['sacro','profano'],['naturale','artificiale'],['silenzio','assordante'],['urlo','silenzioso'],['ghiaccio','bollente'],['fuoco','freddo'],['dolce','amaro'],['innocente','colpevole']
  ];
  const toks=tokenizeWords(text).map(w=>cleanWord(w,false)); const hits=[];
  for(let i=0;i<toks.length;i++) for(let j=i+1;j<Math.min(toks.length,i+5);j++){
    for(const [a,b] of pairs){ if((toks[i]===a&&toks[j]===b)||(toks[i]===b&&toks[j]===a)){ hits.push(toks.slice(i,j+1).join(' ')); break; } }
  }
  return `Ossimori / contrasti candidati: ${hits.length}\n\n`+[...new Set(hits)].join('\n');
}

async function analyze(text,op,p,sensitive){
  const words=tokenizeWords(text), lines=normalizeLines(text);
  switch(op){
    case 0:{ if(!p.trim())return 'Scrivi nel campo Parametro le lettere da contare.'; const seen=new Set(),targets=[]; for(const r of Array.from(p)){if(/\s/u.test(r))continue;const k=normalizeCase(r,sensitive);if(!seen.has(k)){seen.add(k);targets.push(k);}} const counts=new Map(targets.map(x=>[x,0])); for(const r of Array.from(normalizeCase(text,sensitive))) if(seen.has(r))counts.set(r,(counts.get(r)||0)+1); const total=[...counts.values()].reduce((a,b)=>a+b,0); return targets.map(r=>`${r} = ${counts.get(r)||0}`).join('\n')+`\n\nTotale caratteri cercati: ${total}`; }
    case 1:return characterFrequency(text,sensitive);
    case 2:return wordFrequencySummary(words,sensitive);
    case 3: return `Righe: ${lines.length}\nRighe non vuote: ${lines.filter(x=>x.trim()).length}`;
    case 4:{let spaces=0,white=0;for(const r of Array.from(text)){if(r===' ')spaces++;if(/\s/u.test(r))white++;}return `Spazi semplici: ${spaces}\nCaratteri bianchi totali: ${white}`;}
    case 5: return `Cifre: ${(text.match(/\p{N}/gu)||[]).length}`;
    case 6: return `Segni di punteggiatura: ${(text.match(/\p{P}/gu)||[]).length}`;
    case 7:{const q=p.trim();if(!q)return 'Scrivi una parola o stringa da cercare.';return `"${q}"\nOccorrenze: ${countOccurrences(normalizeCase(text,sensitive),normalizeCase(q,sensitive))}`;}
    case 8:{const {freq,display}=freqWords(words,sensitive);const keys=[...freq.keys()].sort();return limitLines(`Parole uniche: ${keys.length}\n\n`+keys.map(k=>display.get(k)).join('\n'),700);}
    case 9:{const {freq,display}=freqWords(words,sensitive);const a=[...freq.entries()].filter(([,n])=>n>1).sort((x,y)=>y[1]-x[1]||x[0].localeCompare(y[0]));return limitLines(`Parole ripetute: ${a.length}\n\n`+a.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'),700);}
    case 10:{const f=new Map();for(const r of Array.from(normalizeCase(text,sensitive)))if(/\p{L}/u.test(r))f.set(r,(f.get(r)||0)+1);return [...f.entries()].sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0])).map(([r,n])=>`${r} = ${n}`).join('\n');}
    case 11:{const {freq,display}=freqWords(words,sensitive);return limitLines([...freq.entries()].sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0])).map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'),700);}
    case 12:{if(!words.length)return 'Nessuna parola.';let min=1e9,max=0,mins=[],maxs=[];for(const w of words){const n=Array.from(cleanWord(w,true)).length;if(!n)continue;if(n<min){min=n;mins=[w]}else if(n===min&&!mins.includes(w))mins.push(w);if(n>max){max=n;maxs=[w]}else if(n===max&&!maxs.includes(w))maxs.push(w);}return `Più corta (${min}): ${mins.join(', ')}\n\nPiù lunga (${max}): ${maxs.join(', ')}`;}
    case 13:return findPalindromes(words,sensitive,parseMinLen(p));
    case 14:return findBifronti(words,sensitive,parseMinLen(p));
    case 15:return await findAnagrams(words,sensitive,parseMinLen(p));
    case 16:{const a=[];for(const line of lines){const t=line.trim();if(!t)continue;const m=t.match(/[\p{L}\p{N}]/u);if(m)a.push(m[0]);}return `Acrostico (${a.length} righe):\n\n${a.join('')}`;}
    case 17:{const a=[];for(const line of lines){const t=line.trim();if(!t)continue;const rr=Array.from(t);for(let i=rr.length-1;i>=0;i--)if(/[\p{L}\p{N}]/u.test(rr[i])){a.push(rr[i]);break;}}return `Telestico (${a.length} righe):\n\n${a.join('')}`;}
    case 18:return reverseRunes(text);
    case 19:return [...words].reverse().join(' ');
    case 20:return [...words].sort((a,b)=>normalizeCase(a,sensitive).localeCompare(normalizeCase(b,sensitive),'it')).join('\n');
    case 21:{const nums=text.match(/\p{N}+/gu)||[];return `Sequenze numeriche: ${nums.length}\n\n${nums.join('\n')}`;}
    case 22:return analyzeRhymeScheme(text,sensitive,parsePositive(p,3,1,12));
    case 23:return findInclusionChains(words,sensitive,parsePositive(p,3,2,20));
    case 24:return findLetterTransport(words,sensitive,parsePositive(p,3,2,20));
    case 25:return findSciarades(text,sensitive,parsePositive(p,2,1,8));
    case 26:return kuntaPhenomena(text,sensitive);
    case 27:return groupWordSignatures(words,'Isovocaliche · stesso scheletro vocalico',vowelSkeleton);
    case 28:return groupWordSignatures(words,'Isoconsonantiche · stesso scheletro consonantico',consonantSkeleton);
    case 29:return groupWordMaterials(words,'Omovocaliche · stesso materiale vocalico',vowelSkeleton);
    case 30:return groupWordMaterials(words,'Omoconsonantiche · stesso materiale consonantico',consonantSkeleton);
    case 31:{const n=parsePositive(p,1,1,8);return groupWordSignatures(words,`Omovocaliche iniziali · prime ${n} vocali`,w=>edgeSignature(vowelSkeleton(w),n,false));}
    case 32:{const n=parsePositive(p,1,1,8);return groupWordSignatures(words,`Omovocaliche finali · ultime ${n} vocali`,w=>edgeSignature(vowelSkeleton(w),n,true));}
    case 33:{const n=parsePositive(p,1,1,8);return groupWordSignatures(words,`Omoconsonantiche iniziali · prime ${n} consonanti`,w=>edgeSignature(consonantSkeleton(w),n,false));}
    case 34:{const n=parsePositive(p,1,1,8);return groupWordSignatures(words,`Omoconsonantiche finali · ultime ${n} consonanti`,w=>edgeSignature(consonantSkeleton(w),n,true));}
    case 35:{if(!words.length)return 'Nessuna parola.';const lens=words.map(wordLen).filter(Boolean);const avg=lens.reduce((a,b)=>a+b,0)/lens.length;return `Parole: ${lens.length}\nLunghezza media: ${avg.toFixed(2)} lettere`;}
    case 36:{const m=new Map();for(const w of words){const n=wordLen(w);if(n)m.set(n,(m.get(n)||0)+1);}return [...m.entries()].sort((a,b)=>a[0]-b[0]).map(([n,c])=>`${n} lettere = ${c}`).join('\n');}
    case 37:return filterWordsBy(words,p,sensitive,'start');
    case 38:return filterWordsBy(words,p,sensitive,'end');
    case 39:return filterWordsBy(words,p,sensitive,'contains');
    case 40:{const n=parsePositive(p,3,1,100);const a=uniqueDisplayWords(words,sensitive).filter(w=>wordLen(w)===n);return `Parole di ${n} lettere: ${a.length}\n\n`+a.join('\n');}
    case 41:return findRepeatedRuns(words);
    case 42:return findRepeatedSequences(words,parsePositive(p,2,2,20));
    case 43:{const a=uniqueDisplayWords(words,false).filter(isIsogram);return `Isogrammi: ${a.length}\n\n`+a.join('\n');}
    case 44:{const a=uniqueDisplayWords(words,false).filter(w=>alphabeticWord(w,false));return `Parole alfabetiche: ${a.length}\n\n`+a.join('\n');}
    case 45:{const a=uniqueDisplayWords(words,false).filter(w=>alphabeticWord(w,true));return `Parole alfabetiche inverse: ${a.length}\n\n`+a.join('\n');}
    case 46:return uniqueDisplayWords(words,sensitive).join('\n');
    case 47:return words.join('\n');
    case 48:return text.toLocaleUpperCase('it');
    case 49:return text.toLocaleLowerCase('it');
    case 50:return findReversePairs(words,sensitive,parsePositive(p,3,2,100),'Palindromi inversi');
    case 51:return findContraryPalindromes(words,parseMinLen(p));
    case 52:return findTransformPairs(words,sensitive,parsePositive(p,3,2,100),'Inversi',transformInverse);
    case 53:return findTransformPairs(words,sensitive,parsePositive(p,3,2,100),'Antipodi',transformAntipode);
    case 54:return groupByKey(words,'Allitterazioni',w=>Array.from(cleanWord(w,false))[0]||'');
    case 55:return groupByKey(words,'Assonanze',w=>suffixChars(vowelSkeleton(w),2));
    case 56:return groupByKey(words,'Consonanze',w=>suffixChars(consonantSkeleton(w),2));
    case 57:return findOxymoronCandidates(text);
    default:return 'Operazione non riconosciuta.';
  }
}


async function readClipboardText(){
  if(window.KuntaNative && typeof window.KuntaNative.readClipboard === 'function') return String(window.KuntaNative.readClipboard() || '');
  if(navigator.clipboard && navigator.clipboard.readText) return navigator.clipboard.readText();
  throw new Error('Clipboard non disponibile');
}
async function writeClipboardText(text){
  if(window.KuntaNative && typeof window.KuntaNative.writeClipboard === 'function'){ window.KuntaNative.writeClipboard(String(text)); return; }
  if(navigator.clipboard && navigator.clipboard.writeText){ await navigator.clipboard.writeText(text); return; }
  throw new Error('Clipboard non disponibile');
}
async function shareText(text){
  if(window.KuntaNative && typeof window.KuntaNative.share === 'function'){ window.KuntaNative.share(String(text)); return; }
  if(navigator.share){ await navigator.share({title:'Kunta',text}); return; }
  throw new Error('Condivisione non disponibile');
}

async function run(){
  if(!input.value){ output.value='Non c’è testo da analizzare.'; return; }
  const op=Number(opSel.value);
  if(op===15){output.value='Caricamento dizionario e ricerca anagrammi…';$('status').textContent='Ricerca anagrammi…';}
  try{ output.value=await analyze(input.value,op,param.value,$('caseSensitive').checked); $('status').textContent='Locale · nessun invio esterno'; }
  catch(err){ output.value='Errore: '+(err&&err.message?err.message:String(err)); $('status').textContent='Errore durante l’analisi.'; }
  output.scrollTop=0;
}
$('runBtn').addEventListener('click',run);
$('.quick');
document.querySelectorAll('button.quick').forEach(b=>b.addEventListener('click',()=>{opSel.value=b.dataset.op;updateHint();run();}));
$('clearBtn').addEventListener('click',()=>{input.value='';output.value='';param.value='';input.focus();});
$('openBtn').addEventListener('click',()=>$('fileInput').click());
$('fileInput').addEventListener('change',async e=>{const f=e.target.files[0];if(!f)return;input.value=await f.text();e.target.value='';});
['dragenter','dragover'].forEach(ev=>document.addEventListener(ev,e=>{e.preventDefault();}));
document.addEventListener('drop',async e=>{e.preventDefault();const f=e.dataTransfer&&e.dataTransfer.files&&e.dataTransfer.files[0];if(f){try{input.value=await f.text();$('status').textContent='Testo caricato.';}catch{}}});
$('pasteBtn').addEventListener('click',async()=>{try{input.value=await readClipboardText();}catch{input.focus();$('status').textContent='Clipboard non disponibile: usa Incolla nel campo testo.';}});
$('copyBtn').addEventListener('click',async()=>{if(!output.value)return;try{await writeClipboardText(output.value);$('status').textContent='Risultato copiato.';}catch{output.select();document.execCommand('copy');}});
$('shareBtn').addEventListener('click',async()=>{if(!output.value)return;try{await shareText(output.value);}catch{$('status').textContent='Condivisione di sistema non disponibile.';}});

const shared=localStorage.getItem('kunta.sharedText');
if(shared){input.value=shared;localStorage.removeItem('kunta.sharedText');$('status').textContent='Testo ricevuto dalla condivisione.';}

let installPrompt=null;
window.addEventListener('beforeinstallprompt',e=>{e.preventDefault();installPrompt=e;$('installBtn').classList.remove('hidden');});
$('installBtn').addEventListener('click',async()=>{if(!installPrompt)return;installPrompt.prompt();await installPrompt.userChoice;installPrompt=null;$('installBtn').classList.add('hidden');});
window.addEventListener('appinstalled',()=>{$('status').textContent='Kunta installato nella schermata Home.';});

if('serviceWorker' in navigator && location.protocol.startsWith('http')) navigator.serviceWorker.register('./sw.js');
