import llms from '../../archive/llms.txt?raw';

export function GET() {
  return new Response(llms, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
