#!/usr/bin/env python3

import PyPDF2
import os

def extract_pdf_text(pdf_path):
    try:
        with open(pdf_path, 'rb') as file:
            pdf_reader = PyPDF2.PdfReader(file)
            text = ""
            for page_num in range(len(pdf_reader.pages)):
                page = pdf_reader.pages[page_num]
                text += f"\n--- Page {page_num + 1} ---\n"
                text += page.extract_text()
            return text
    except Exception as e:
        return f"Error reading {pdf_path}: {str(e)}"

# List of PDF files to read
pdf_files = [
    "/Users/bhagya/work/personal/engramiq/Supporting-Documents/Sample-Reports.pdf",
    "/Users/bhagya/work/personal/engramiq/Supporting-Documents/STYLE-GUIDE.pdf", 
    "/Users/bhagya/work/personal/engramiq/Engramiq_Tech_lead_assignment.pdf",
    "/Users/bhagya/work/personal/engramiq/Engramiq__PRD_RAwidget.pdf"
]

for pdf_file in pdf_files:
    print(f"\n{'='*80}")
    print(f"EXTRACTING: {os.path.basename(pdf_file)}")
    print(f"{'='*80}")
    
    if os.path.exists(pdf_file):
        content = extract_pdf_text(pdf_file)
        print(content)
    else:
        print(f"File not found: {pdf_file}")
    
    print(f"\n{'='*80}")
    print(f"END OF: {os.path.basename(pdf_file)}")
    print(f"{'='*80}\n")