from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.enums import TA_RIGHT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import mm
from reportlab.platypus import KeepTogether, ListFlowable, ListItem, Paragraph, SimpleDocTemplate, Spacer, Table, TableStyle


OUT = Path(__file__).resolve().parents[1] / 'public' / 'Jordan-Cooper-CV.pdf'
ACCENT = colors.HexColor('#c47a2a')
INK = colors.HexColor('#1a1a1a')
MUTED = colors.HexColor('#70685f')
RULE = colors.HexColor('#ddd5ca')

achievements = [
    "Built and led Zumi's engineering organisation from scratch, growing it to 10 engineers and contractors and establishing hiring, coaching, delivery, security and engineering standards.",
    'Designed reliable production AI capabilities including agents, custom MCP servers, secure user-scoped API access, evaluation frameworks and observability.',
    'Designed backend services and distributed systems supporting OAuth 2.0 authentication and authorisation, Open Banking, payment initiation and secure financial data access.',
    'Delivered AI capabilities and multi-tenant SaaS products from concept to production.',
    "Built Open Banking and payment products used by more than 1.6 million UK consumers, including one of the UK's first white-labelled Open Banking payment widgets.",
    'Worked with founders, VC investors, enterprise customers and distributed teams across the UK, US, EU and Middle East to shape product and engineering direction.',
]

zumi = [
    'Joined as employee #1 and built the engineering function, platform architecture and engineering operating model from zero.',
    'Built and led a 10-person engineering team, with responsibility for hiring, coaching, career development, performance and delivery.',
    'Worked with founders, enterprise customers and distributed teams across the UK, US and Middle East through multiple company pivots spanning hospitality, payments, AI and data.',
    'Designed backend services, APIs and multi-tenant SaaS platforms using TypeScript, PostgreSQL, Terraform, Docker and Google Cloud.',
    'Led production AI development including agents, custom MCP servers, LLM middleware, secure user-scoped API access, request tracking, observability, evaluations and multi-model routing.',
    'Built integrations with payments, CRM, loyalty, reservations, PMS and POS systems.',
    'Led multiple products from concept to production, balancing customer needs, technical trade-offs and long-term platform investment.',
    'Established engineering practices covering architecture, code review, CI/CD, incident response, operational excellence and production reliability.',
    'Partnered with founders, product leaders, investors and enterprise customers on roadmaps, architecture and technical strategy.',
]

moneyhub = [
    'Progressed through four engineering leadership roles during a period of rapid company growth.',
    'Led engineering teams building payments, mobile applications and Open Banking capabilities supporting more than 1.6 million UK consumers.',
    'Managed engineers across multiple teams, supporting hiring, coaching, technical direction and performance development.',
    'Designed backend services, APIs and distributed systems supporting OAuth 2.0 authentication and authorisation, payment initiation and secure financial data access.',
    "Led development of one of the UK's first white-labelled Open Banking payment widgets for major financial institutions.",
    'Worked with product, compliance, commercial and executive stakeholders to deliver secure, regulated financial products with strong governance and compliance requirements.',
    'Balanced customer needs, regulatory requirements and long-term technical strategy whilst delivering production systems at scale.',
]


def build_pdf():
    OUT.parent.mkdir(parents=True, exist_ok=True)
    doc = SimpleDocTemplate(str(OUT), pagesize=A4, rightMargin=16 * mm, leftMargin=16 * mm, topMargin=15 * mm, bottomMargin=14 * mm)
    styles = getSampleStyleSheet()
    body = ParagraphStyle('Body', parent=styles['BodyText'], fontName='Helvetica', fontSize=8.6, leading=12, textColor=INK, spaceAfter=5)
    heading = ParagraphStyle('Heading', parent=body, fontName='Helvetica-Bold', fontSize=8, leading=10, textColor=MUTED, spaceBefore=11, spaceAfter=6)
    name = ParagraphStyle('Name', parent=body, fontName='Helvetica-Bold', fontSize=22, leading=24, textColor=INK, spaceAfter=3)
    title = ParagraphStyle('Title', parent=body, fontName='Helvetica-Bold', fontSize=9.5, leading=12, textColor=INK, spaceAfter=4)
    meta = ParagraphStyle('Meta', parent=body, fontName='Helvetica', fontSize=8, leading=11, textColor=MUTED)
    job = ParagraphStyle('Job', parent=body, fontName='Helvetica-Bold', fontSize=9.2, leading=11, textColor=INK)
    dates = ParagraphStyle('Dates', parent=meta, alignment=TA_RIGHT)
    bullet = ParagraphStyle('Bullet', parent=body, leftIndent=10, firstLineIndent=-7, spaceAfter=2)

    story = [
        Paragraph('Jordan Cooper', name),
        Paragraph('Engineering Leader | AI Agents | Backend &amp; Platform Engineering', title),
        Paragraph('London Area, UK | Open to Remote &amp; Hybrid Opportunities', meta),
        Paragraph('jordancooper@hey.com | jordanjoecooper.com | github.com/jordanjoecooper | linkedin.com/in/jordanjoecooper', meta),
        Spacer(1, 8),
        Paragraph('Summary', heading),
    ]
    for text in [
        'Engineering leader with 10+ years of experience building backend platforms, customer-facing products and engineering teams across fintech, payments and AI.',
        'Most recently Head of Engineering at Zumi, joining as employee #1 to build the engineering organisation from the ground up whilst remaining hands-on with architecture and delivery. Led production AI work spanning agents, custom Model Context Protocol (MCP) servers, secure user-scoped API integrations and multi-tenant SaaS platforms.',
        'Previously at Moneyhub, I built Open Banking APIs and payment products used by more than 1.6 million UK consumers, progressing through Software Engineer, Tech Lead, Engineering Manager and Solutions Architect roles.',
        'I enjoy building high-performing teams, staying close to the technology and developing reliable platforms that balance product delivery, long-term engineering investment and enterprise trust.',
    ]:
        story.append(Paragraph(text, body))

    story += [Paragraph('Selected Achievements', heading), bullets(achievements, bullet), Paragraph('Experience', heading)]
    story += job_block('Zumi (DVx Ventures)', 'Head of Engineering', 'Mar 2024 - Jul 2026', zumi, job, dates, meta, bullet)
    story += job_block('Moneyhub', 'Software Engineer | Solutions Architect | Tech Lead / EM', '2020 - 2024', moneyhub, job, dates, meta, bullet)
    story += [Paragraph('Earlier Experience', heading)]
    story.append(Paragraph('<b>Nimble</b> - Software Developer (2018-2020). Built frontend applications, backend services, APIs and integrations for an e-learning platform.', body))
    story.append(Paragraph('<b>Kappu</b> - Software Developer &amp; Data Protection Officer (2017-2018). First engineering hire. Built the platform alongside the founders whilst establishing engineering processes and delivery practices.', body))
    story += [Paragraph('Core Expertise', heading), Paragraph('Engineering Leadership | AI Platforms | AI Agents | MCP | Backend &amp; Platform Engineering | Distributed Systems | OAuth 2.0 | APIs &amp; Enterprise Integrations | Open Banking | Multi-tenant SaaS | TypeScript | PostgreSQL | GCP | Terraform', body), Paragraph('Education', heading), Paragraph('<b>BSc (Hons) Computing (2:1), FdSc Computing</b><br/>Canterbury Christ Church University', body)]
    doc.build(story, onFirstPage=footer, onLaterPages=footer)


def bullets(items, style):
    return ListFlowable([ListItem(Paragraph(item, style), leftIndent=0) for item in items], bulletType='bullet', start='-', leftIndent=10, bulletFontName='Helvetica', bulletFontSize=6, bulletColor=ACCENT, spaceAfter=4)


def job_block(company, role, dates_text, items, job_style, dates_style, meta_style, bullet_style):
    header = Table([[Paragraph(company, job_style), Paragraph(dates_text, dates_style)]], colWidths=[125 * mm, 52 * mm])
    header.setStyle(TableStyle([('VALIGN', (0, 0), (-1, -1), 'TOP'), ('LEFTPADDING', (0, 0), (-1, -1), 0), ('RIGHTPADDING', (0, 0), (-1, -1), 0), ('TOPPADDING', (0, 0), (-1, -1), 0), ('BOTTOMPADDING', (0, 0), (-1, -1), 0)]))
    return [KeepTogether([header, Paragraph(role, meta_style), Spacer(1, 2)]), bullets(items, bullet_style), Spacer(1, 5)]


def footer(canvas, doc):
    canvas.saveState()
    canvas.setStrokeColor(RULE)
    canvas.setLineWidth(0.5)
    canvas.line(doc.leftMargin, 9 * mm, A4[0] - doc.rightMargin, 9 * mm)
    canvas.setFillColor(MUTED)
    canvas.setFont('Helvetica', 7)
    canvas.drawString(doc.leftMargin, 5.5 * mm, 'Jordan Cooper')
    canvas.drawRightString(A4[0] - doc.rightMargin, 5.5 * mm, f'Page {doc.page}')
    canvas.restoreState()


if __name__ == '__main__':
    build_pdf()
