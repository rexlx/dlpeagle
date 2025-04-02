import pikepdf
import sys
import os

def add_url_action(pdf_path, output_path, page_number, x, y, width, height, url):
    with pikepdf.open(pdf_path) as pdf:
        page = pdf.pages[page_number]

        # Create the annotation dictionary directly.
        annot = pikepdf.Dictionary(
            Subtype=pikepdf.Name('/Link'),
            Rect=[x, y, x + width, y + height],
            A=pikepdf.Dictionary(S=pikepdf.Name('/URI'), URI=url)
        )

        # Append the dictionary to the page's annotations.
        if "/Annots" in page:
            page["/Annots"].append(annot)
        else:
            page["/Annots"] = pikepdf.Array([annot])

        pdf.save(output_path)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python script.py <input_pdf_path>")
        sys.exit(1)

    pdf_path = sys.argv[1]
    output_path = os.path.splitext(pdf_path)[0] + "_new.pdf"
    page_number = 0  # Page index (0-based)
    x, y, width, height = 100, 700, 200, 50  # Coordinates and size of the clickable area
    url = "http://fairlady:8081/okay"

    add_url_action(pdf_path, output_path, page_number, x, y, width, height, url)
    print(f"Modified PDF saved to: {output_path}")