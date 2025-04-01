import pikepdf

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

# Example usage:
pdf_path = "/Users/rxlx/Documents/resume/rFitzhugh.pdf"
output_path = "output.pdf"
page_number = 0  # Page index (0-based)
x, y, width, height = 100, 700, 200, 50  # Coordinates and size of the clickable area
url = "http://fairlady:8081/okay"

add_url_action(pdf_path, output_path, page_number, x, y, width, height, url)
