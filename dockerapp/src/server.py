# This Code provided a RESTful access to the functions of the SentenceTransfer Modules.
# The calling code (written in golang / will call out to this code via HTTP call. 

from sentence_transformers import SentenceTransformer, util
import hug
from hug_middleware_cors import CORSMiddleware


@hug.post("/sim")
def sim(reference: str, text: str):
    """Get entities for displaCy ENT visualizer."""
    
    embeddings1 = model.encode(reference, convert_to_tensor=True)
    embeddings2 = model.encode(text, convert_to_tensor=True)

    cosine_scores = util.pytorch_cos_sim(embeddings1, embeddings2)

    return { "similarty" : float(cosine_scores[0][0])}




# Model used is the MSMacro version - msmarco-distilbert-base-?? 
# on first run / the model will be downloaded from HuggingFace (there can be a small delay while this is happening) 

#https://public.ukp.informatik.tu-darmstadt.de/reimers/sentence-transformers/v0.2/

#great results returned for initial inputs using the -v3
#model = SentenceTransformer('msmarco-distilbert-base-v3') # used for initial paper
model = SentenceTransformer('msmarco-distilbert-base-v4') 

if __name__ == "__main__":
    import waitress

# start - Example 
# Two lists of sentences for an example of the similatiy (using non / REST - also proves )
sentences1 = ['Read forms',
             'Read forms',
             'The new movie is awesome']

sentences2 = ['Position construction forms or molds',
              'Prepare forms or applications',
              'The new movie is so great']

#Compute embedding for both lists
embeddings1 = model.encode(sentences1, convert_to_tensor=True)
embeddings2 = model.encode(sentences2, convert_to_tensor=True)

#Compute cosine-similarits
cosine_scores = util.pytorch_cos_sim(embeddings1, embeddings2)

#Output the pairs with their score
for i in range(len(sentences1)):
    print("{} \t\t {} \t\t Score: {:.4f}".format(sentences1[i], sentences2[i], cosine_scores[i][i]))

# end - Example 

    app = hug.API(__name__)
    app.http.add_middleware(CORSMiddleware(app))
    waitress.serve(__hug_wsgi__, port=8083) # REST endpoint opened on PORT 8083

