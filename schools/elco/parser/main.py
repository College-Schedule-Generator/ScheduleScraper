import json
from pathlib import Path
from pprint import pprint
from typing import List, TypedDict

from bs4 import BeautifulSoup


class CourseType(TypedDict):
    name: str
    year: int

courses: List[CourseType] = []			

for file in Path('../cache/2022-SU').glob('*.html'):
	with open(file) as f:
		content = f.read()
		soup = BeautifulSoup(content, 'html.parser')
		tbl = soup.find('tbody', { 'class': 'esg-table-body' })
		for tblRow in tbl.find_all('tr', { 'recursive': False }):
			termCell, statusCell, sectionNameCell, titleCell, datesCell, locationCell, instructionalMethodsCell, meetingInformationCell, facultyCell, availabilityCell, creditsCell, academicLevelCell, commentsCell, bookStoreCell, *_ = tblRow.find_all('td', { 'recursive': False })
			courses.append({
				'term': termCell.text.strip(),
				'status': statusCell.text.strip(),
				'sectionName': sectionNameCell.text.strip(),
				'title': titleCell.text.strip(),
				'dates': datesCell.text.strip(),
				'location': locationCell.text.strip(),
				'instructionalMethods': instructionalMethodsCell.text.strip(),
				'meetingInformation': meetingInformationCell.text.strip(),
				'faculty': facultyCell.text.strip(),
				'availability': availabilityCell.text.strip(),
				'credits': creditsCell.text.strip(),
				'academicLevel': academicLevelCell.text.strip(),
				'comments': commentsCell.text.strip(),
				'bookStore': bookStoreCell.text.strip()
			})


outFile = '../output/courses.json'
Path(Path(outFile).parent).mkdir(parents=True, exist_ok=False)
Path(outFile).write_text(json.dumps({
	'courses': courses
}))
