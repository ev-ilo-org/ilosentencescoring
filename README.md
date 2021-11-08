# ilosentencescoring
 ILO Sentence Scoring Service


**Pre-Requirements (e.g. MACOS)**
- Docker Desktop 
- Install DOCKER Desktop from https://www.docker.com/get-started

- Golang compiler 
- Install Golang https://golang.org/doc/install



1. Clone the repo/or download a zip to your computer
2. Build the Docker image (this install all the various opensource modules etc + source code) **(can be skipped if making use of East Village Server deployment)**
3. Build execuable command line utility 
4. Run command line utility

![builddockerscreenshot](https://user-images.githubusercontent.com/11387813/140821719-b97e1a99-8b82-4c8a-a5f4-032cad562f28.png)


**Building the Docker Image**
- Navigate to the folder **dockerapp** 
- processed to run the shell command

In terminal (shell) 
cd dockerapp
../builddockerimage.sh
(Wait for a short while while the image is build. Docker will start running the image as soon as it's succesfully built)


![dockerimage](https://user-images.githubusercontent.com/11387813/140822400-2a70cdc9-6215-4484-bdeb-d09d492d0e3c.png)

To test if the server is responding using the following curl command

**curl --location --request POST 'localhost:8083/sim' \
--header 'Content-Type: application/json' \
--data-raw '{
   "reference": "Read instructions",
   "text": "Read manuals"
}'**

*Note should localhost not be found / replace with 127.0.0.1

To check if using East Village Server

**curl --location --request POST 'ilo.eastvillagescl.com:8083/sim' \
--header 'Content-Type: application/json' \
--data-raw '{
   "reference": "Read instructions",
   "text": "Read manuals"
}'**



**Building the command line utility**

go build -o ilo main.go  _replace ilo with the name you require_


**running the command line utility**

move to the **referenceinputs** subdirectory
cd referenceinputs

run
**../ilo -ref reference_sentencesPIAAC.csv -input DWA_cat.csv -server 127.0.0.1 -max=2**



arguments
<ref> reference sentence inputfile .csv [ default **reference_sentencesPIAAC.csv** ]

**Format is CSV file (example) **
 
1,working cooperatively or collaboratively with co-workers
2,sharing work-related information with co-workers
3,"instructing, training or teaching people, individually or in groups"
4,making speeches or giving presentations in front of five or more people
5,selling a product or selling a service
 
 
<input> file for processing [default **DWA_cat.csv** ]
 
 
**Format is CSV file (example) **
 
Element Name,IWA Title,DWA Title
Getting Information,Study details of artistic productions.,Review art or design materials.
Getting Information,Study details of artistic productions.,Study details of musical compositions.
Getting Information,Study details of artistic productions.,Review production information to determine costume or makeup requirements.
Getting Information,Study details of artistic productions.,Study scripts to determine project requirements.
Getting Information,Read documents or materials to inform work processes.,Read materials to determine needed actions.
Getting Information,Read documents or materials to inform work processes.,Read maps to determine routes.
Getting Information,Read documents or materials to inform work processes.,Review customer information.


 <output>
 Used to set the name of the output results.csv file [ default **results.csv** ]
  
 
<server> 
**localhost or 127.0.0.1 **(when running docker or running the python server code directly)
**ilo.eastvillagescl.com **(when using the server from eastvillage software consultants) [**default**]
 
<max> [ **default = process everything** ]
Allows a limit on the number of sentences processed e.g. setting max=2 / just two lines will be processed - this allowing to check the results
 


